package auth

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"account-service/service/internal/logging"

	"github.com/creack/pty"
	gomessage "github.com/gif-gif/go.io/go-message"
	"github.com/gif-gif/go.io/go-utils/gojson"
	"github.com/rs/zerolog"
)

var Kiro = KiroCli{}

// 容器内的授权之后的地址
const authTokenPath = "/app/.aws/sso/cache/kiro-auth-token-cli.json"

var kiroLoginURLPattern = regexp.MustCompile(`https://app\.kiro\.dev/account/device\?user_code=[A-Z0-9-]+&login_provider=[a-zA-Z]+`)
var kiroAWSLoginURLPattern = regexp.MustCompile(`Open this URL:\s*(https://[^\s]+/#/device\?user_code=[A-Z0-9-]+)`)

func kiroLogger() zerolog.Logger {
	return logging.Default().With().Str("component", "kiro_login").Logger()
}

type kiroAuthStatus string

const (
	kiroAuthPending   kiroAuthStatus = "pending"
	kiroAuthSucceeded kiroAuthStatus = "succeeded"
	kiroAuthFailed    kiroAuthStatus = "failed"
)

type KiroCli struct {
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
	feishuWebhook string
}

type KiroCliAccount struct {
	ID          string
	Username    string
	LoginURL    string
	Region      string
	AccountType string
}

type KiroCliConfig struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	ProfileArn   string    `json:"profileArn"`
	AuthMethod   string    `json:"authMethod"`
	Provider     string    `json:"provider"`
}

func (k *KiroCli) Running() bool {
	return k.running
}

func (k *KiroCli) SetFeishuWebhook(webhook string) {
	k.feishuWebhook = strings.TrimSpace(webhook)
}

func (k *KiroCli) Cancel() {
	if k.cancel != nil {
		k.cancel()
	}
}

// 不支持并发操作，只有一个容器
func (k *KiroCli) KiroCliLogin(account KiroCliAccount) (bool, *KiroCliConfig) {
	logger := kiroLogger()
	if k.running {
		k.running = false
		if k.cancel != nil {
			k.cancel()
		}
		return false, nil
	}
	k.running = true
	k.ctx, k.cancel = context.WithCancel(context.Background())
	defer func() {
		k.running = false
		k.cancel()
	}()
	urlChan := make(chan string, 1)

	if k.feishuWebhook != "" {
		go func(webhook string) {
			for url := range urlChan {
				gomessage.FeiShu(webhook, fmt.Sprintf("\n🚀 [外部接口收到通知] 拿到目标 URL: %s", url))
			}
		}(k.feishuWebhook)
	}

	success := kiroCliLogin(k.ctx, urlChan)
	cliConfig := &KiroCliConfig{}
	if success {
		err := gojson.UnmarshalFromFile(authTokenPath, cliConfig)
		if err != nil {
			logger.Error().Err(err).Msg("Kiro-CLI 获取配置解析失败")
			return false, nil
		}
		logger.Info().Msg("Kiro-CLI 登录成功")
	} else {
		logger.Info().Msg("Kiro-CLI 登录失败、超时或被外部取消")
	}
	return success, cliConfig
}

// lineBuffer 持续从 reader 读取行，存入缓冲区，支持并发安全读取
type lineBuffer struct {
	mu    sync.Mutex
	lines []string
	all   string
}

func (lb *lineBuffer) append(line string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.lines = append(lb.lines, line)
	lb.all += line + "\n"
}

func (lb *lineBuffer) getAll() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.all
}

func (lb *lineBuffer) contains(s string) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return strings.Contains(lb.all, s)
}

func kiroAWSLoginArgs(account KiroCliAccount) []string {
	return []string{
		"login",
		"--license",
		"pro",
		"--identity-provider",
		kiroAWSIdentityProvider(account),
		"--region",
		kiroAWSRegion(account),
	}
}

func kiroAWSIdentityProvider(account KiroCliAccount) string {
	if strings.TrimSpace(account.LoginURL) == "" {
		return "https://d-90660ed8222222.awsapps.com/start"
	}
	return strings.TrimSpace(account.LoginURL)
}

func kiroAWSRegion(account KiroCliAccount) string {
	if strings.TrimSpace(account.Region) == "" {
		return "us-east-1"
	}
	return strings.TrimSpace(account.Region)
}

func extractKiroAWSLoginURL(output string) string {
	match := kiroAWSLoginURLPattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func kiroCliLogin(ctx context.Context, urlChan chan string) bool {
	defer close(urlChan)
	logger := kiroLogger()

	if err := ctx.Err(); err != nil {
		logger.Error().Err(err).Msg("流程启动前已被取消")
		return false
	}

	// ================== 前置登出逻辑 ==================
	logger.Info().Msg("开始执行 logout")
	logoutCmd := exec.Command("kiro-cli", "logout")
	logoutCmd.Stdin = nil
	logoutCmd.Stdout = nil
	logoutCmd.Stderr = nil
	logoutCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := logoutCmd.Start(); err != nil {
		logger.Error().Err(err).Msg("logout Start 失败")
	} else {
		logger.Info().Int("pid", logoutCmd.Process.Pid).Msg("logout 已启动")
		done := make(chan error, 1)
		go func() { done <- logoutCmd.Wait() }()
		select {
		case err := <-done:
			logger.Info().Err(err).Msg("logout 正常退出")
		case <-time.After(3 * time.Second):
			logger.Warn().Msg("logout 3秒超时，强杀")
			_ = syscall.Kill(-logoutCmd.Process.Pid, syscall.SIGKILL)
			_ = logoutCmd.Process.Kill()
			<-done
			logger.Warn().Msg("logout 已强杀")
		}
	}
	logger.Info().Msg("logout 阶段结束")

	// ================== 使用 creack/pty 启动进程 ==================
	logger.Info().Msg("正在通过 pty 启动 kiro-cli login")
	cmd := exec.CommandContext(ctx, "kiro-cli", "login", "--use-device-flow")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		logger.Error().Err(err).Msg("pty.Start 失败")
		return false
	}
	defer ptmx.Close()
	logger.Info().Int("pid", cmd.Process.Pid).Msg("kiro-cli login 已启动")

	// 后台持续读取 pty 输出
	output := &lineBuffer{}
	newLine := make(chan string, 100)

	go func() {
		scanner := bufio.NewScanner(ptmx)
		for scanner.Scan() {
			line := scanner.Text()
			output.append(line)
			select {
			case newLine <- line:
			default:
			}
		}
		close(newLine)
	}()

	// ================== 阶段一：选择登录方式 ==================
	logger.Info().Msg("等待 Select login method 提示")
	if !waitForString(ctx, output, newLine, "Select login method", 10*time.Second) {
		logger.Error().Str("output", truncate(output.getAll(), 500)).Msg("等不到选择提示")
		killProcess(cmd)
		return false
	}
	logger.Info().Msg("收到 Select login method，发送向下键和回车")

	time.Sleep(300 * time.Millisecond)
	_, _ = ptmx.Write([]byte("\x1b[B"))
	time.Sleep(300 * time.Millisecond)
	_, _ = ptmx.Write([]byte("\r"))
	logger.Info().Msg("已发送选择操作")

	// ================== 阶段二：捕获 URL ==================
	logger.Info().Msg("开始捕获登录 URL")
	reURL := kiroLoginURLPattern //regexp.MustCompile(`https://app\.kiro\.dev/account/device\?user_code=[A-Z0-9-]+&login_provider=[a-zA-Z]+`)

	targetURL := ""
	urlDeadline := time.After(20 * time.Second)
urlLoop:
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("在等待 URL 时流程被外部取消")
			killProcess(cmd)
			return false
		case <-urlDeadline:
			logger.Error().Str("output", truncate(output.getAll(), 800)).Msg("未能在终端里找到登录 URL")
			killProcess(cmd)
			return false
		case _, ok := <-newLine:
			if !ok {
				logger.Error().Msg("进程输出已关闭，未找到 URL")
				killProcess(cmd)
				return false
			}
			all := output.getAll()
			if match := reURL.FindString(all); match != "" {
				targetURL = match
				break urlLoop
			}
		case <-time.After(200 * time.Millisecond):
			all := output.getAll()
			if match := reURL.FindString(all); match != "" {
				targetURL = match
				break urlLoop
			}
		}
	}

	logger.Info().Str("target_url", targetURL).Msg("成功提取 URL")

	select {
	case urlChan <- targetURL:
		logger.Info().Msg("URL 已发送到 channel")
	case <-ctx.Done():
		logger.Info().Msg("在发送 URL 时流程被外部取消")
		killProcess(cmd)
		return false
	}

	// ================== 阶段三：等待授权 (2分钟超时 + Context 监听) ==================
	authTimeout := 2 * time.Minute
	logger.Info().Dur("timeout", authTimeout).Msg("进入阶段三：等待授权")

	authDeadline := time.After(authTimeout)
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("收到取消信号，外部主动取消了登录流程")
			killProcess(cmd)
			return false
		case <-authDeadline:
			logger.Warn().Dur("timeout", authTimeout).Msg("授权超时，用户未完成操作")
			killProcess(cmd)
			return false
		case _, ok := <-newLine:
			if !ok {
				all := output.getAll()
				if strings.Contains(all, "Signed in with Google") {
					logger.Info().Msg("登录成功，检测到 Signed in with Google")
					_ = cmd.Wait()
					return true
				}
				logger.Error().Msg("进程已退出，未检测到登录成功")
				return false
			}
			all := output.getAll()
			if strings.Contains(all, "Signed in with Google") {
				logger.Info().Msg("登录成功，检测到 Signed in with Google")
				_ = cmd.Wait()
				return true
			}
			if matched, _ := regexp.MatchString(`(error|expired)`, all); matched {
				logger.Error().Msg("授权失败，终端检测到验证码过期或授权错误 outputAll :" + all)
				killProcess(cmd)
				return false
			}
		case <-time.After(2 * time.Second):
			all := output.getAll()
			if strings.Contains(all, "Signed in with Google") {
				logger.Info().Msg("登录成功，检测到 Signed in with Google")
				_ = cmd.Wait()
				return true
			}
		}
	}
}

func (k *KiroCli) KiroCliLoginByAws(account KiroCliAccount) (bool, *KiroCliConfig) {
	logger := kiroLogger()
	if k.running {
		k.running = false
		if k.cancel != nil {
			k.cancel()
		}
		return false, nil
	}
	k.running = true
	k.ctx, k.cancel = context.WithCancel(context.Background())
	defer func() {
		k.running = false
		k.cancel()
	}()
	urlChan := make(chan string, 1)

	if k.feishuWebhook != "" {
		go func(webhook string) {
			for url := range urlChan {
				gomessage.FeiShu(webhook, fmt.Sprintf("\n🚀 [外部接口收到通知] 拿到目标 URL: %s", url))
			}
		}(k.feishuWebhook)
	}

	success := kiroCliLoginByAws(k.ctx, urlChan, account)
	cliConfig := &KiroCliConfig{}
	if success {
		err := gojson.UnmarshalFromFile(authTokenPath, cliConfig)
		if err != nil {
			logger.Error().Err(err).Msg("Kiro-CLI AWS 获取配置解析失败")
			return false, nil
		}
		logger.Info().Msg("Kiro-CLI AWS 登录成功")
	} else {
		logger.Info().Msg("Kiro-CLI AWS 登录失败、超时或被外部取消")
	}
	return success, cliConfig
}

func kiroCliLoginByAws(ctx context.Context, urlChan chan string, account KiroCliAccount) bool {
	defer close(urlChan)
	logger := kiroLogger()

	if err := ctx.Err(); err != nil {
		logger.Error().Err(err).Msg("AWS 登录流程启动前已被取消")
		return false
	}

	logger.Info().Msg("开始执行 logout")
	logoutCmd := exec.Command("kiro-cli", "logout")
	logoutCmd.Stdin = nil
	logoutCmd.Stdout = nil
	logoutCmd.Stderr = nil
	logoutCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := logoutCmd.Start(); err != nil {
		logger.Error().Err(err).Msg("logout Start 失败")
	} else {
		logger.Info().Int("pid", logoutCmd.Process.Pid).Msg("logout 已启动")
		done := make(chan error, 1)
		go func() { done <- logoutCmd.Wait() }()
		select {
		case err := <-done:
			logger.Info().Err(err).Msg("logout 正常退出")
		case <-time.After(3 * time.Second):
			logger.Warn().Msg("logout 3秒超时，强杀")
			_ = syscall.Kill(-logoutCmd.Process.Pid, syscall.SIGKILL)
			_ = logoutCmd.Process.Kill()
			<-done
			logger.Warn().Msg("logout 已强杀")
		}
	}
	logger.Info().Msg("logout 阶段结束")

	logger.Info().Msg("正在通过 pty 启动 kiro-cli AWS 登录")
	cmd := exec.CommandContext(ctx, "kiro-cli", kiroAWSLoginArgs(account)...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		logger.Error().Err(err).Msg("pty.Start 失败")
		return false
	}
	defer ptmx.Close()
	logger.Info().Int("pid", cmd.Process.Pid).Msg("kiro-cli AWS 登录已启动")

	output := &lineBuffer{}
	newLine := make(chan string, 100)

	go func() {
		scanner := bufio.NewScanner(ptmx)
		for scanner.Scan() {
			line := scanner.Text()
			output.append(line)
			select {
			case newLine <- line:
			default:
			}
		}
		close(newLine)
	}()

	logger.Info().Msg("开始捕获 AWS 登录 URL")
	targetURL := ""
	urlDeadline := time.After(20 * time.Second)
urlLoop:
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("在等待 AWS URL 时流程被外部取消")
			killProcess(cmd)
			return false
		case <-urlDeadline:
			logger.Error().Str("output", truncate(output.getAll(), 800)).Msg("未能在终端里找到 AWS 登录 URL")
			killProcess(cmd)
			return false
		case _, ok := <-newLine:
			if !ok {
				logger.Error().Msg("进程输出已关闭，未找到 AWS 登录 URL")
				killProcess(cmd)
				return false
			}
			if match := extractKiroAWSLoginURL(output.getAll()); match != "" {
				targetURL = match
				break urlLoop
			}
		case <-time.After(200 * time.Millisecond):
			if match := extractKiroAWSLoginURL(output.getAll()); match != "" {
				targetURL = match
				break urlLoop
			}
		}
	}

	logger.Info().Str("target_url", targetURL).Msg("成功提取 AWS 登录 URL")

	select {
	case urlChan <- targetURL:
		logger.Info().Msg("AWS URL 已发送到 channel")
	case <-ctx.Done():
		logger.Info().Msg("在发送 AWS URL 时流程被外部取消")
		killProcess(cmd)
		return false
	}

	authTimeout := 2 * time.Minute
	logger.Info().Dur("timeout", authTimeout).Msg("进入 AWS 登录授权等待")

	authDeadline := time.After(authTimeout)
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("收到取消信号，外部主动取消了 AWS 登录流程")
			killProcess(cmd)
			return false
		case <-authDeadline:
			logger.Warn().Dur("timeout", authTimeout).Msg("AWS 授权超时，用户未完成操作")
			killProcess(cmd)
			return false
		case _, ok := <-newLine:
			all := output.getAll()
			if strings.Contains(all, "Logged in successfully") || strings.Contains(all, "Signed in with Google") {
				logger.Info().Msg("AWS 登录成功")
				_ = cmd.Wait()
				return true
			}
			if matched, _ := regexp.MatchString(`(?i)(error|expired|failed)`, all); matched {
				logger.Error().Str("output", truncate(all, 800)).Msg("AWS 授权失败，终端检测到错误")
				killProcess(cmd)
				return false
			}
			if !ok {
				logger.Error().Str("output", truncate(all, 800)).Msg("进程已退出，未检测到 AWS 登录成功")
				return false
			}
		case <-time.After(2 * time.Second):
			all := output.getAll()
			if strings.Contains(all, "Logged in successfully") || strings.Contains(all, "Signed in with Google") {
				logger.Info().Msg("AWS 登录成功")
				_ = cmd.Wait()
				return true
			}
		}
	}
}

func waitForString(ctx context.Context, output *lineBuffer, newLine <-chan string, target string, timeout time.Duration) bool {
	deadline := time.After(timeout)
	for {
		if output.contains(target) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-deadline:
			return false
		case _, ok := <-newLine:
			if !ok {
				return output.contains(target)
			}
			if output.contains(target) {
				return true
			}
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func killProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	logger := kiroLogger()
	logger.Info().Int("pid", cmd.Process.Pid).Msg("killProcess")
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// 确保 io 包被使用（bufio.Scanner 间接依赖）
var _ = io.EOF
