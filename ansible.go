package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

type HostConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type HostsFile struct {
	Hosts []HostConfig `yaml:"hosts"`
}

func main() {
	hostsFile := flag.String("f", "hosts.yaml", "hosts file")
	cmdFile := flag.String("c", "", "command file (multi-line commands)")
	parallel := flag.Int("p", 5, "parallel workers")
	logEnable := flag.Bool("log", false, "enable logging to ./logs")
	flag.Parse()

	// 如果没有 -c，则取最后一个参数作为命令
	var cmds []string
	cmdName := "cmd" // 用于日志文件名
	if *cmdFile != "" {
		cmds = loadCommandsFromFile(*cmdFile)
		cmdName = filepath.Base(*cmdFile)
	} else {
		args := flag.Args()
		if len(args) == 0 {
			log.Fatal("no command provided")
		}
		cmds = []string{args[len(args)-1]}
	}

	hosts := loadHosts(*hostsFile)

	// 确保 logs 目录存在
	if *logEnable {
		os.MkdirAll("./logs", 0755)
	}

	sem := make(chan struct{}, *parallel)
	var wg sync.WaitGroup

	for _, h := range hosts {
		wg.Add(1)
		go func(host HostConfig) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			runHost(host, cmds, *logEnable, cmdName)
		}(h)
	}

	wg.Wait()
}

/* ---------------- hosts.yaml ---------------- */

func loadHosts(path string) []HostConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var cfg HostsFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}
	return cfg.Hosts
}

/* ---------------- command loader ---------------- */

func loadCommandsFromFile(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return splitLines(string(data))
}

func splitLines(s string) []string {
	var cmds []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		cmds = append(cmds, l)
	}
	return cmds
}

/* ---------------- ssh + executor ---------------- */

func runHost(h HostConfig, cmds []string, logEnable bool, cmdName string) {
	info, err := loadSSH(h.Host, h.Port)
	if err != nil {
		printHost(h.Host, fmt.Sprintf("ssh config error: %v", err), logEnable, cmdName)
		return
	}

	client, err := ssh.Dial("tcp", info.addr, &ssh.ClientConfig{
		User:            info.user,
		Auth:            []ssh.AuthMethod{info.auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		printHost(h.Host, fmt.Sprintf("ssh connect error: %v", err), logEnable, cmdName)
		return
	}
	defer client.Close()

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("===== [%s] =====\n", h.Host))

	for _, cmd := range cmds {
		buf.WriteString(fmt.Sprintf("$ %s\n", cmd))
		out, err := runSudoBuffered(client, h.Password, cmd)
		if err != nil {
			buf.WriteString(fmt.Sprintf("ERROR: %v\n", err))
			break
		}
		if out != "" {
			buf.WriteString(out + "\n")
		}
	}

	printHost(h.Host, buf.String(), logEnable, cmdName)
}

var printLock sync.Mutex

func printHost(host, output string, logEnable bool, cmdName string) {
	printLock.Lock()
	defer printLock.Unlock()

	fmt.Println(output)

	if logEnable {
		ts := time.Now().Format("20060102-150405")
		fileName := fmt.Sprintf("./logs/%s-%s-%s.log", host, cmdName, ts)
		os.WriteFile(fileName, []byte(output), 0644)
	}
}

func runSudoBuffered(client *ssh.Client, password, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	in, _ := session.StdinPipe()
	cmd := fmt.Sprintf("sudo -S bash -c %q", command)

	if err := session.Start(cmd); err != nil {
		return "", err
	}

	fmt.Fprintln(in, password)

	if err := session.Wait(); err != nil {
		return "", fmt.Errorf("%s", stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

/* ---------------- ssh config ---------------- */

type sshInfo struct {
	user string
	addr string
	auth ssh.AuthMethod
}

func loadSSH(host string, overridePort int) (*sshInfo, error) {
	user := ssh_config.Get(host, "User")
	hostname := ssh_config.Get(host, "HostName")
	port := ssh_config.Get(host, "Port")
	key := ssh_config.Get(host, "IdentityFile")

	if hostname == "" {
		return nil, fmt.Errorf("host %s not found in ssh config", host)
	}

	if overridePort != 0 {
		port = fmt.Sprintf("%d", overridePort)
	}
	if port == "" {
		port = "22"
	}

	keyPath := expandPath(key)
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, err
	}

	return &sshInfo{
		user: user,
		addr: hostname + ":" + port,
		auth: ssh.PublicKeys(signer),
	}, nil
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}
