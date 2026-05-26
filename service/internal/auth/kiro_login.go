package auth

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"account-service/service/internal/logging"

	"github.com/creack/pty"
	gomessage "github.com/gif-gif/go.io/go-message"
	"github.com/gif-gif/go.io/go-utils/gojson"
)

var Kiro = KiroCli{}

const authTokenPath = "~/.aws/sso/cache/kiro-auth-token-cli.json"

var kiroLoginURLPattern = regexp.MustCompile(`https://view\.awsapps\.com/start/#/device\?user_code=[A-Z0-9-]+`)

type kiroAuthStatus string

const (
	kiroAuthPending   kiroAuthStatus = "pending"
	kiroAuthSucceeded kiroAuthStatus = "succeeded"
	kiroAuthFailed    kiroAuthStatus = "failed"
)

type KiroCli struct {
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

type KiroCliConfig struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	ProfileArn   string    `json:"profileArn"`
	AuthMethod   string    `json:"authMethod"`
	Provider     string    `json:"provider"`
}

func (k *KiroCli) Cancel() {
	if k.cancel != nil {
		k.cancel()
	}
}

func (k *KiroCli) KiroCliLogin() (bool, *KiroCliConfig) {
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

	go func() {
		for url := range urlChan {
			gomessage.FeiShu("https://open.feishu.cn/open-apis/bot/v2/hook/82c21496-edd0-43c1-a848-945139307a64", fmt.Sprintf("\n🚀 [外部接口收到通知] 拿到目标 URL: %s", url))
		}
	}()

	success := kiroCliLogin(k.ctx, urlChan)
	cliConfig := &KiroCliConfig{}
	if success {
		err := gojson.UnmarshalFromFile(authTokenPath, cliConfig)
		if err != nil {
			logKiroError("Kiro-CLI 获取配置解析失败", err)
			return false, nil
		}
		logKiroInfo("Kiro-CLI 登录成功")
	} else {
		logKiroInfo("Kiro-CLI 登录失败、超时或被外部取消")
	}

	return success, cliConfig
}

func kiroCliLogin(ctx context.Context, urlChan chan string) bool {
	defer close(urlChan)

	if err := ctx.Err(); err != nil {
		logKiroError("流程启动前已被取消", err)
		return false
	}

	// ================== 前置登出逻辑 ==================
	logKiroInfo("开始执行 logout")
	logoutCmd := exec.Command("kiro-cli", "logout")
	logoutCmd.Stdin = nil
	logoutCmd.Stdout = nil
	logoutCmd.Stderr = nil
	logoutCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	logKiroInfo("logout 命令已创建，准备 Start")
	if err := logoutCmd.Start(); err != nil {
		logKiroError("logout Start 失败", err)
	} else {
		logKiroInfof("logout 已启动, PID=%d, 等待完成(3秒超时)", logoutCmd.Process.Pid)
		done := make(chan error, 1)
		go func() { done <- logoutCmd.Wait() }()
		select {
		case err := <-done:
			logKiroInfof("logout 正常退出, err=%v", err)
		case <-time.After(3 * time.Second):
			logKiroInfo("logout 3秒超时，发送 SIGKILL")
			_ = syscall.Kill(-logoutCmd.Process.Pid, syscall.SIGKILL)
			_ = logoutCmd.Process.Kill()
			<-done
			logKiroInfo("logout 已强杀")
		}
	}
	logKiroInfo("logout 阶段结束")

	// ================== 启动进程 ==================
	logKiroInfo("正在启动 kiro-cli login")
	cmd := kiroLoginCommand()
	cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		logKiroError("启动 kiro-cli 失败", err)
		return false
	}
	defer ptmx.Close()
	logKiroInfof("kiro-cli login 已启动, PID=%d", cmd.Process.Pid)

	output := strings.Builder{}
	chunks := readPTYOutput(ptmx)

	// ================== 阶段一：捕获 URL ==================
	logKiroInfo("开始捕获登录 URL")
	logKiroInfo("进入循环读取终端输出(15秒超时)")
	text, err := waitForPTYOutput(ctx, chunks, 15*time.Second, &output, func(text string) (bool, error) {
		return extractKiroLoginURL(text) != "", nil
	})
	if err != nil {
		logKiroInfof("未能在终端里找到登录 URL, buf 长度=%d", output.Len())
		logKiroInfof("buf 内容(前500字符): %s", truncate(output.String(), 500))
		killProcess(cmd)
		return false
	}
	logKiroInfo("匹配到 URL")

	targetURL := extractKiroLoginURL(text)
	if targetURL == "" {
		logKiroInfo("未能从终端输出中提取登录 URL")
		killProcess(cmd)
		return false
	}
	logKiroInfof("成功提取 URL: %s", targetURL)

	select {
	case urlChan <- targetURL:
		logKiroInfo("URL 已发送到 channel")
	case <-ctx.Done():
		logKiroInfo("在发送 URL 时流程被外部取消")
		killProcess(cmd)
		return false
	}

	// ================== 阶段二：等待授权 (2分钟超时 + Context 监听) ==================
	authTimeout := 2 * time.Minute
	logKiroInfof("进入阶段二：等待授权 (超时: %v)", authTimeout)

	text, err = waitForPTYOutput(ctx, chunks, authTimeout, &output, func(text string) (bool, error) {
		switch kiroAuthOutputStatus(text) {
		case kiroAuthSucceeded:
			return true, nil
		case kiroAuthFailed:
			return false, fmt.Errorf("auth failed")
		default:
			return false, nil
		}
	})
	if err != nil {
		if strings.Contains(err.Error(), "auth failed") {
			logKiroInfo("授权失败，终端检测到验证码过期或授权错误")
		} else if ctx.Err() != nil {
			logKiroInfo("收到取消信号，外部主动取消了登录流程，正在强杀命令行")
		} else {
			logKiroInfof("授权超时，用户在 %v 内未完成操作", authTimeout)
		}
		killProcess(cmd)
		return false
	}

	if kiroAuthOutputStatus(text) == kiroAuthSucceeded {
		logKiroInfo("登录成功，检测到 Logged in successfully")
		_ = cmd.Wait()
		return true
	}

	killProcess(cmd)
	return false
}

func kiroLoginCommand() *exec.Cmd {
	return exec.Command("kiro-cli", "login", "--use-device-flow")
}

func killProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	logKiroInfof("killProcess PID=%d", cmd.Process.Pid)
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

type ptyChunk struct {
	text string
	err  error
}

func readPTYOutput(reader io.Reader) <-chan ptyChunk {
	chunks := make(chan ptyChunk, 16)
	go func() {
		defer close(chunks)
		buf := make([]byte, 4096)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				chunks <- ptyChunk{text: string(buf[:n])}
			}
			if err != nil {
				chunks <- ptyChunk{err: err}
				return
			}
		}
	}()
	return chunks
}

func waitForPTYOutput(
	ctx context.Context,
	chunks <-chan ptyChunk,
	timeout time.Duration,
	accumulated *strings.Builder,
	match func(string) (bool, error),
) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		if ok, err := match(accumulated.String()); ok || err != nil {
			return accumulated.String(), err
		}

		select {
		case <-ctx.Done():
			return accumulated.String(), ctx.Err()
		case <-timer.C:
			return accumulated.String(), fmt.Errorf("timeout")
		case chunk, ok := <-chunks:
			if !ok {
				return accumulated.String(), io.EOF
			}
			if chunk.text != "" {
				accumulated.WriteString(chunk.text)
				logKiroInfof("读取到数据, 当前 buf 长度=%d", accumulated.Len())
			}
			if chunk.err != nil {
				return accumulated.String(), fmt.Errorf("%w; output=%q", chunk.err, truncate(accumulated.String(), 500))
			}
		}
	}
}

func extractKiroLoginURL(output string) string {
	return kiroLoginURLPattern.FindString(output)
}

func kiroAuthOutputStatus(output string) kiroAuthStatus {
	if strings.Contains(output, "Signed in with Google") || strings.Contains(output, "Logged in successfully") {
		return kiroAuthSucceeded
	}
	if matched, _ := regexp.MatchString(`(?i)(error|expired)`, output); matched {
		return kiroAuthFailed
	}
	return kiroAuthPending
}

func logKiroInfo(message string) {
	logger := logging.Default()
	logger.Info().Str("component", "kiro_login").Msg(message)
}

func logKiroInfof(format string, args ...any) {
	logKiroInfo(fmt.Sprintf(format, args...))
}

func logKiroError(message string, err error) {
	logger := logging.Default()
	logger.Error().Err(err).Str("component", "kiro_login").Msg(message)
}
