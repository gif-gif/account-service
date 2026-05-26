package auth

import "testing"

func TestExtractKiroLoginURL(t *testing.T) {
	output := "open https://app.kiro.dev/account/device?user_code=ABCD-1234&login_provider=google to continue"

	url := extractKiroLoginURL(output)

	if url != "https://app.kiro.dev/account/device?user_code=ABCD-1234&login_provider=google" {
		t.Fatalf("url = %q", url)
	}
}

func TestKiroAuthOutputStatus(t *testing.T) {
	if status := kiroAuthOutputStatus("Signed in with Google"); status != kiroAuthSucceeded {
		t.Fatalf("status = %q, want success", status)
	}
	if status := kiroAuthOutputStatus("device code expired"); status != kiroAuthFailed {
		t.Fatalf("status = %q, want failed", status)
	}
	if status := kiroAuthOutputStatus("still waiting"); status != kiroAuthPending {
		t.Fatalf("status = %q, want pending", status)
	}
}
