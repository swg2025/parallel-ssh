# Ansible-like SSH Command Execution Tool

This is a simple Go tool for executing SSH commands in parallel across multiple servers, similar to Ansible's functionality.

## Features

- **Parallel Execution**: Run commands simultaneously across multiple hosts
- **SSH Configuration Integration**: Uses system SSH configuration files (~/.ssh/config)
- **Multiple Command Input**: Supports single commands or batch reading from command files
- **Logging**: Option to save output to log files
- **Password Authentication**: Supports sudo command execution (requires password)

## Requirements

- Go 1.16 or higher
- SSH keys configured on target hosts

## Installation

```bash
go build ansible.go
```

## Usage

### Basic Syntax

```bash
./ansible [options] [command]
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-f` | [hosts.yaml](file:///Users/huosi/dev/go/go-admin/tools/hosts.yaml) | Hosts configuration file path |
| `-c` | None | Command file path (multi-line commands) |
| `-p` | `5` | Number of parallel workers |
| `-log` | `false` | Enable logging to ./logs directory |

### Examples

#### 1. Execute a single command
```bash
./ansible -f my_hosts.yaml -p 10 "uptime"
```

#### 2. Execute all commands from a command file
```bash
./ansible -f my_hosts.yaml -c commands.txt -p 5 -log
```

#### 3. Execute command on default hosts.yaml
```bash
./ansible "df -h"
```

### Configuration Files

#### hosts.yaml format
```yaml
hosts:
  - host: server1.example.com
    port: 22
    password: your_sudo_password
  - host: server2.example.com
    port: 22
    password: your_sudo_password
```

#### Command file format
```
# Comments starting with # are ignored
ls -la /tmp
df -h
free -m
whoami
```

### SSH Configuration

This tool automatically reads SSH configuration from the SSH config file (typically ~/.ssh/config):

```
Host server1.example.com
    User username
    HostName 192.168.1.100
    Port 22
    IdentityFile ~/.ssh/id_rsa
```

## Output

- Console displays execution results for each host
- If logging is enabled, each host's output is saved to a separate log file
- Log file naming format: `./logs/{hostname}-{cmdname}-{timestamp}.log`

## Notes

- Ensure SSH keys are properly configured on target hosts
- All hosts must be accessible via SSH
- Provided password is only used for sudo privilege escalation
- Comments in command files start with `#`

## Error Handling

- If SSH connection fails, error information will be displayed in the corresponding host's output
- If command execution fails, error information will be shown and the next command will continue
- If host configuration file doesn't exist, the program will terminate

## Chinese Version

[中文版文档](Readme-zhcn.md)