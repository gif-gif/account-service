package auth

import "testing"

func TestKiroCliSetFeishuWebhookTrimsValue(t *testing.T) {
	kiro := KiroCli{}

	kiro.SetFeishuWebhook(" https://open.feishu.cn/open-apis/bot/v2/hook/example ")

	want := "https://open.feishu.cn/open-apis/bot/v2/hook/example"
	if kiro.feishuWebhook != want {
		t.Fatalf("feishuWebhook = %q, want %q", kiro.feishuWebhook, want)
	}
}

func TestKiroLoginCommandUsesDeviceFlow(t *testing.T) {

}

func TestExtractKiroLoginURL(t *testing.T) {

}

func TestKiroAuthOutputStatus(t *testing.T) {

}
