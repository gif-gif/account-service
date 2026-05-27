package auth

import (
	"bytes"
	"testing"
)

func TestKiroCliTargetURLStoresLatestValue(t *testing.T) {
	kiro := KiroCli{}

	kiro.setTargetURL(" https://app.kiro.dev/account/device?user_code=ABCD-EFGH&login_provider=google ")

	want := "https://app.kiro.dev/account/device?user_code=ABCD-EFGH&login_provider=google"
	if kiro.TargetURL() != want {
		t.Fatalf("TargetURL() = %q, want %q", kiro.TargetURL(), want)
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

func TestSendKiroAWSLoginDefaultAnswersWritesTwoEnters(t *testing.T) {
	var out bytes.Buffer

	if err := sendKiroAWSLoginDefaultAnswers(&out); err != nil {
		t.Fatalf("sendKiroAWSLoginDefaultAnswers() error = %v", err)
	}

	if got := out.String(); got != "\r\r" {
		t.Fatalf("written input = %q, want two carriage returns", got)
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
