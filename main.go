package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 全局变量：改为存储多用户信息（map结构，key=用户名，value=密码）
var (
	users    map[string]string // 多用户列表
	rootPath string
	port     int
	logPath  string
)

// getClientIP 提取客户端真实IP地址（去掉端口）
func getClientIP(r *http.Request) string {
	ipPort := r.RemoteAddr
	ip, _, err := net.SplitHostPort(ipPort)
	if err != nil {
		return ipPort
	}
	return ip
}

// initLogger 初始化日志配置：同时输出到控制台和指定文件（如果设置了日志路径）
func initLogger() error {
	if logPath == "" {
		return nil
	}

	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	return nil
}

// parseUsers 解析多用户参数（格式：user1:pass1,user2:pass2）
func parseUsers(userStr string) map[string]string {
	userMap := make(map[string]string)
	if userStr == "" {
		// 默认用户
		userMap["admin"] = "123456"
		return userMap
	}

	// 分割多个用户
	userPairs := strings.Split(userStr, ",")
	for _, pair := range userPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		// 分割用户名和密码
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			log.Fatalf("用户参数格式错误：%s（正确格式：user:pass,user2:pass2）", pair)
		}
		username := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])
		if username == "" || password == "" {
			log.Fatalf("用户名/密码不能为空：%s", pair)
		}
		userMap[username] = password
	}

	if len(userMap) == 0 {
		log.Fatal("未配置有效用户，请检查-users参数")
	}
	return userMap
}

// basicAuth 中间件：支持多用户认证
func basicAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		user, pass, ok := r.BasicAuth()

		if !ok {
			// 未提供认证信息
			log.Printf("[%s] 认证失败 - 未提供用户名密码 (请求路径: %s)", clientIP, r.URL.Path)
			w.Header().Set("WWW-Authenticate", `Basic realm="File Server Login"`)
			http.Error(w, "认证失败，请输入正确的用户名和密码", http.StatusUnauthorized)
			return
		}

		// 校验多用户
		expectedPass, exists := users[user]
		if !exists || pass != expectedPass {
			log.Printf("[%s] 认证失败 - 用户名/密码错误 (用户名: %s, 请求路径: %s)", clientIP, user, r.URL.Path)
			w.Header().Set("WWW-Authenticate", `Basic realm="File Server Login"`)
			http.Error(w, "认证失败，请输入正确的用户名和密码", http.StatusUnauthorized)
			return
		}

		// 认证成功
		log.Printf("[%s] 认证成功 - 用户名: %s (请求路径: %s)", clientIP, user, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}

// loggingFileHandler 包装文件处理器，记录文件访问日志
func loggingFileHandler(root http.FileSystem) http.Handler {
	fileServer := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		log.Printf("[%s] 文件访问 - 路径: %s", clientIP, r.URL.Path)
		fileServer.ServeHTTP(w, r)
	})
}

func main() {
	// 定义启动参数：替换原有的-user/-pass为-users
	var userStr string
	flag.IntVar(&port, "port", 8080, "服务监听的端口号，默认8080")
	flag.StringVar(&rootPath, "path", "./", "要共享的本地文件路径，默认当前目录")
	flag.StringVar(&userStr, "users", "", "多用户配置（格式：user1:pass1,user2:pass2），默认 admin:123456")
	flag.StringVar(&logPath, "logpath", "", "日志文件存放路径（可选），如 /var/log/file-server.log")
	flag.Parse()

	// 初始化日志
	if err := initLogger(); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}

	// 解析多用户
	users = parseUsers(userStr)

	// 校验共享路径
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		log.Fatalf("路径解析失败：%v", err)
	}
	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		log.Fatalf("指定的共享路径不存在：%s", absPath)
	}
	if err != nil {
		log.Fatalf("路径访问失败：%v", err)
	}

	// 创建文件服务器
	fileServer := loggingFileHandler(http.Dir(absPath))
	http.Handle("/", basicAuth(fileServer))

	// 启动信息
	addr := fmt.Sprintf(":%d", port)
	log.Printf("===== 文件服务器启动信息 =====")
	log.Printf("访问地址：http://localhost%s", addr)
	log.Printf("共享路径：%s", absPath)
	log.Printf("配置的用户列表：%v", users)
	if logPath != "" {
		log.Printf("日志文件路径：%s", logPath)
	}
	log.Printf("==============================")
	log.Printf("（按Ctrl+C停止服务）")

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("服务启动失败：%v", err)
	}
}