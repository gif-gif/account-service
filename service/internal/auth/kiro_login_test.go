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

func TestKiroAWSLoginCommandUsesLicenseAndIdentityProvider(t *testing.T) {
	args := kiroAWSLoginArgs(KiroCliAccount{
		LoginURL: "https://d-90660ed825.awsapps.com/start",
		Region:   "eu-west-1",
	})

	want := []string{
		"login",
		"--use-device-flow",
		"--license",
		"pro",
		"--identity-provider",
		"https://d-90660ed825.awsapps.com/start",
		"--region",
		"eu-west-1",
	}
	if len(args) != len(want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestExtractKiroAWSLoginURL(t *testing.T) {
	output := `✔ Enter Start URL · https://d-90660ed825.awsapps.com/start
✔ Enter Region · us-east-1

Confirm the following code in the browser
Code: MPPG-MKGV

Open this URL: https://d-90660ed825.awsapps.com/start/#/device?user_code=MPPG-MKGV
▰▰▰▰▰▰▱ Logging in..`

	got := extractKiroAWSLoginURL(output)

	want := "https://d-90660ed825.awsapps.com/start/#/device?user_code=MPPG-MKGV"
	if got != want {
		t.Fatalf("extractKiroAWSLoginURL() = %q, want %q", got, want)
	}
}

func TestExtractKiroLoginURL(t *testing.T) {
}

func TestKiroAuthOutputStatus(t *testing.T) {

}
