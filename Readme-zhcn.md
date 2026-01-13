# Ansible-like SSH Command Execution Tool

这是一个简单的 Go 工具，用于在多台服务器上并行执行 SSH 命令，类似于 Ansible 的功能。

## 功能特性

- **并行执行**：可同时在多个主机上运行命令
- **SSH 配置集成**：使用系统 SSH 配置文件（~/.ssh/config）
- **多种命令输入方式**：支持单个命令或从文件批量读取命令
- **日志记录**：可选择将输出保存到日志文件
- **密码认证**：支持 sudo 命令执行（需要提供密码）

## 安装要求

- Go 1.16 或更高版本
- SSH 密钥已配置在目标主机上

## 安装方法

```bash
go build ansible.go
```

## 使用方法

### 基本语法

```bash
./ansible [options] [command]
```

### 参数说明

| 参数 | 默认值 | 描述 |
|------|--------|------|
| `-f` | [hosts.yaml](file:///Users/huosi/dev/go/go-admin/tools/hosts.yaml) | 指定主机配置文件路径 |
| `-c` | 无 | 指定包含多行命令的文件路径 |
| `-p` | `5` | 并行工作进程数量 |
| `-log` | `false` | 启用日志记录到 ./logs 目录 |

### 示例

#### 1. 执行单个命令
```bash
./ansible -f my_hosts.yaml -p 10 "uptime"
```

#### 2. 执行命令文件中的所有命令
```bash
./ansible -f my_hosts.yaml -c commands.txt -p 5 -log
```

#### 3. 在默认 hosts.yaml 上执行命令
```bash
./ansible "df -h"
```

### 配置文件

#### hosts.yaml 格式
```yaml
hosts:
  - host: server1.example.com
    port: 22
    password: your_sudo_password
  - host: server2.example.com
    port: 22
    password: your_sudo_password
```

#### 命令文件格式
```
# 这是注释行会被忽略
ls -la /tmp
df -h
free -m
whoami
```

### SSH 配置

此工具会自动从 SSH 配置文件（通常为 ~/.ssh/config）中读取以下设置：

```
Host server1.example.com
    User username
    HostName 192.168.1.100
    Port 22
    IdentityFile ~/.ssh/id_rsa
```

## 输出

- 控制台会显示每个主机的执行结果
- 如果启用日志记录，每个主机的输出会保存到独立的日志文件中
- 日志文件命名格式：`./logs/{hostname}-{cmdname}-{timestamp}.log`

## 注意事项

- 确保 SSH 密钥已在目标主机上正确配置
- 所有主机必须能通过 SSH 访问
- 提供的密码仅用于 sudo 权限提升
- 命令文件中的注释以 `#` 开头

## 错误处理

- 如果 SSH 连接失败，会在相应主机的输出中显示错误信息
- 如果命令执行失败，会显示错误信息并继续执行下一个命令
- 如果主机配置文件不存在，程序会终止执行

