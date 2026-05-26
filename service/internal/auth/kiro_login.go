package auth

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"account-service/service/internal/logging"

	"github.com/Netflix/go-expect"
	gomessage "github.com/gif-gif/go.io/go-message"
	"github.com/gif-gif/go.io/go-utils/gojson"
)

var Kiro = KiroCli{}

const authTokenPath = "~/.aws/sso/cache/kiro-auth-token-cli.json"

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

	// ================== 初始化虚拟终端 ==================
	logKiroInfo("正在创建虚拟终端")
	console, err := expect.NewConsole(expect.WithDefaultTimeout(10 * time.Second))
	if err != nil {
		logKiroError("无法初始化 Console", err)
		return false
	}
	defer console.Close()
	logKiroInfo("虚拟终端创建成功")

	// ================== 启动进程 ==================
	logKiroInfo("正在启动 kiro-cli login")
	cmd := exec.CommandContext(ctx, "kiro-cli", "login")
	cmd.Stdin = console.Tty()
	cmd.Stdout = console.Tty()
	cmd.Stderr = console.Tty()

	if err := cmd.Start(); err != nil {
		logKiroError("启动 kiro-cli 失败", err)
		return false
	}
	logKiroInfof("kiro-cli login 已启动, PID=%d", cmd.Process.Pid)

	_ = console.Tty().Close()
	logKiroInfo("tty 已关闭，通过 master 端读写")

	// ================== 阶段一：选择登录方式 ==================
	logKiroInfo("等待 Select login method 提示")
	_, err = console.ExpectString("Select login method")
	if err != nil {
		logKiroError("等不到选择提示或流程被取消", err)
		killProcess(cmd)
		return false
	}
	logKiroInfo("收到 Select login method，发送向下键+回车")

	_, _ = console.Send("\x1b[B")
	time.Sleep(200 * time.Millisecond)
	_, _ = console.Send("\n")
	logKiroInfo("已发送选择操作")

	// ================== 阶段二：捕获 URL ==================
	logKiroInfo("开始捕获登录 URL")
	reURL := regexp.MustCompile(`https://app\.kiro\.dev/account/device\?user_code=[A-Z0-9-]+&login_provider=[a-zA-Z]+`)

	// 用循环读取代替 ExpectEOF，因为进程不会退出
	logKiroInfo("进入循环读取终端输出(15秒超时)")
	deadline := time.After(15 * time.Second)
	var buf string
	for {
		select {
		case <-ctx.Done():
			logKiroInfo("在等待 URL 时流程被外部取消")
			killProcess(cmd)
			return false
		case <-deadline:
			logKiroInfof("未能在终端里找到登录 URL, buf 长度=%d", len(buf))
			logKiroInfof("buf 内容(前500字符): %s", truncate(buf, 500))
			killProcess(cmd)
			return false
		default:
			time.Sleep(100 * time.Millisecond)
			chunk, _ := console.ExpectString("\n")
			if chunk != "" {
				buf += chunk
				logKiroInfof("读取到数据, 当前 buf 长度=%d", len(buf))
				if reURL.MatchString(buf) {
					logKiroInfo("匹配到 URL")
					goto urlFound
				}
			}
		}
	}

urlFound:
	targetURL := reURL.FindString(buf)
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

	// ================== 阶段三：等待授权 (2分钟超时 + Context 监听) ==================
	authTimeout := 2 * time.Minute
	logKiroInfof("进入阶段三：等待授权 (超时: %v)", authTimeout)

	type expectResult struct {
		output string
		err    error
	}

	resultChan := make(chan expectResult, 1)

	go func() {
		deadline := time.Now().Add(authTimeout)
		var accumulated string
		for time.Now().Before(deadline) {
			chunk, err := console.ExpectString("\n")
			accumulated += chunk
			if strings.Contains(accumulated, "Signed in with Google") {
				resultChan <- expectResult{output: accumulated, err: nil}
				return
			}
			if matched, _ := regexp.MatchString(`(error|expired)`, accumulated); matched {
				resultChan <- expectResult{output: accumulated, err: fmt.Errorf("auth failed")}
				return
			}
			if err != nil {
				resultChan <- expectResult{output: accumulated, err: err}
				return
			}
		}
		resultChan <- expectResult{output: accumulated, err: fmt.Errorf("timeout")}
	}()

	select {
	case <-ctx.Done():
		logKiroInfo("收到取消信号，外部主动取消了登录流程，正在强杀命令行")
		killProcess(cmd)
		return false

	case res := <-resultChan:
		if res.err != nil {
			if strings.Contains(res.err.Error(), "auth failed") {
				logKiroInfo("授权失败，终端检测到验证码过期或授权错误")
			} else {
				logKiroInfof("授权超时，用户在 %v 内未完成操作", authTimeout)
			}
			killProcess(cmd)
			return false
		}

		if strings.Contains(res.output, "Signed in with Google") {
			logKiroInfo("登录成功，检测到 Signed in with Google")
			_ = cmd.Wait()
			return true
		}

		killProcess(cmd)
		return false
	}
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
