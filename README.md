# file-to-web
一个轻量级Go语言实现的Linux文件共享Web服务器，支持多用户认证、自定义端口/共享路径/日志路径，记录访问者IP，打包后可直接在Linux运行。

## 功能特性
- ✅ 多用户认证：支持配置多个用户名密码，满足不同人员访问需求
- ✅ 自定义启动参数：端口、共享路径、日志路径均可通过启动参数配置
- ✅ 访问日志：记录访问者IP、认证状态、访问路径，支持日志文件持久化
- ✅ 跨平台编译：Windows/macOS可编译出Linux可执行文件，无依赖
- ✅ 权限校验：启动时校验共享路径/日志路径权限，避免运行时报错

## 环境要求
### 编译环境
- Go 1.18+（Windows/Linux/macOS均可编译）
### 运行环境
- Linux（x86_64/amd64、arm64等架构，推荐x86_64）

## 快速开始
### 1. 获取代码/编译文件
#### 方式1：从GitHub克隆代码编译
```bash
# 克隆仓库
git clone https://github.com/GosbyQin/file-to-web.git
cd file-to-web

# Linux本地编译
go build -ldflags="-s -w" -o file-server main.go

# Windows跨平台编译Linux版本（PowerShell）
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o file-server main.go
```

#### 方式2：直接下载编译产物
从GitHub Releases下载对应架构的`file-server`可执行文件，上传到Linux服务器。

### 2. 赋予执行权限
```bash
chmod +x ./file-server
```

### 3. 启动服务
#### 基础启动（默认配置）
```bash
./file-server
```
默认配置：
- 端口：8080
- 共享路径：当前目录（./）
- 默认用户：admin / 123456
- 日志：仅输出到控制台

#### 自定义参数启动（推荐）
```bash
# 示例：指定端口+共享路径+多用户+日志路径
./file-server \
  -port 8999 \
  -path /yitu/gosby \
  -users "user1:123456,user2:Yitu@123,ops:Ops&789" \
  -logpath /home/yituadmin/gosby/file-to-web/file-server.log
```

#### 后台运行（不中断）
```bash
# nohup后台运行，关闭终端不停止
nohup ./file-server \
  -port 8999 \
  -path /yitu/gosby \
  -users "user1:123456,user2:Yitu@123" \
  -logpath /home/yituadmin/gosby/file-to-web/file-server.log > /dev/null 2>&1 &
```

### 4. 访问服务
1. 打开浏览器，访问 `http://Linux服务器IP:端口`（如 `http://192.168.1.100:8999`）
2. 输入配置的用户名密码，即可浏览/下载共享路径下的文件

## 启动参数说明
| 参数名   | 类型   | 默认值    | 说明                                                                 |
|----------|--------|-----------|----------------------------------------------------------------------|
| `-port`  | int    | 8080      | 服务监听端口（1-65535，1-1024端口需root权限启动）|
| `-path`  | string | ./        | 要共享的本地文件路径（绝对路径/相对路径均可）|
| `-users` | string | admin:123456 | 多用户配置，格式：`用户名1:密码1,用户名2:密码2`，支持特殊字符（建议用引号包裹） |
| `-logpath`| string | 空        | 日志文件存放路径（可选），如 `/var/log/file-server.log`，未指定则仅输出到控制台 |

### 多用户配置示例
| 配置命令                          | 可用用户                     | 说明                     |
|-----------------------------------|------------------------------|--------------------------|
| `-users "test:Test@123"`          | test / Test@123              | 单个自定义用户           |
| `-users "dev:Dev123,ops:Ops456"`  | dev/Dev123、ops/Ops456       | 多个用户，逗号分隔       |
| `-users "admin:Admin@666,guest:123"` | admin/Admin@666、guest/123 | 包含特殊字符的密码       |

## 日志说明
### 日志内容
日志包含以下关键信息，便于审计和问题排查：
- 服务启动/停止信息
- 认证日志：访问IP、用户名、认证成功/失败、请求路径
- 文件访问日志：访问IP、访问的文件/目录路径
- 错误日志：端口占用、路径不存在、权限不足等

### 日志示例
```
2026/02/12 10:00:00 ===== 文件服务器启动信息 =====
2026/02/12 10:00:00 访问地址：http://localhost:8999
2026/02/12 10:00:00 共享路径：/yitu/gosby
2026/02/12 10:00:00 配置的用户列表：map[user1:123456 user2:Yitu@123]
2026/02/12 10:00:00 日志文件路径：/home/yituadmin/gosby/file-to-web/file-server.log
2026/02/12 10:01:20 [192.168.1.105] 认证成功 - 用户名: user1 (请求路径: /)
2026/02/12 10:01:20 [192.168.1.105] 文件访问 - 路径: /
2026/02/12 10:02:15 [10.0.0.5] 认证失败 - 用户名/密码错误 (用户名: test, 请求路径: /)
```

### 日志轮转（可选）
长期运行建议配置`logrotate`避免日志文件过大：
1. 创建配置文件：
   ```bash
   sudo vi /etc/logrotate.d/file-server
   ```
2. 添加以下内容：
   ```
   /home/yituadmin/gosby/file-to-web/file-server.log {
       daily       # 按天轮转
       rotate 7    # 保留7天日志
       compress    # 压缩旧日志
       missingok   # 日志文件不存在时不报错
       notifempty  # 空文件不轮转
       create 0644 root root  # 新建日志文件权限
   }
   ```

## 常见问题
### 1. 编译后Linux无法执行（Exec format error）
- 原因：Windows编译时未设置`GOOS=linux`，生成了Windows可执行文件
- 解决：Windows PowerShell中重新编译：
  ```powershell
  $env:GOOS="linux"
  $env:GOARCH="amd64"
  go build -ldflags="-s -w" -o file-server main.go
  ```

### 2. 端口被占用（bind: address already in use）
- 解决：指定未被占用的端口，如 `-port 8999`

### 3. 权限不足（Permission denied）
- 日志路径/共享路径无写入/访问权限：用`sudo`启动
  ```bash
  sudo ./file-server -logpath /var/log/file-server.log -path /yitu/gosby
  ```
- 1-1024端口需要root权限：`sudo ./file-server -port 80`

### 4. 多用户配置报错（格式错误）
- 原因：未按`用户名:密码`格式配置，或缺少引号
- 解决：确保格式正确，特殊字符用引号包裹：
  ```bash
  ./file-server -users "user1:123456,user2:Yitu@123"
  ```

### 5. Git推送失败（Connection was reset）
- 原因：国内网络访问GitHub不稳定
- 解决：配置GitHub镜像：
  ```powershell
  git config --global url."https://mirror.ghproxy.com/https://github.com/".insteadOf "https://github.com/"
  ```

## 停止服务
```bash
# 查找进程ID
ps -ef | grep file-server | grep -v grep

# 终止进程（替换为实际PID）
kill -9 12345
```

## 版本更新
| 版本 | 变更说明 |
|------|----------|
| v1.0.0 | 基础版本：单用户、自定义端口/路径、日志记录 |
| v1.1.0 | 新增：多用户认证、优化日志格式、完善错误处理 |
```
