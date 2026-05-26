package auth

import "testing"

func TestExtractKiroLoginURL(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "kiro app device url",
			output: "open https://app.kiro.dev/account/device?user_code=ABCD-1234&login_provider=google to continue",
			want:   "https://app.kiro.dev/account/device?user_code=ABCD-1234&login_provider=google",
		},
		{
			name:   "aws device flow url",
			output: "Open this URL: https://view.awsapps.com/start/#/device?user_code=HRMW-GJGH\r\n",
			want:   "https://view.awsapps.com/start/#/device?user_code=HRMW-GJGH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := extractKiroLoginURL(tt.output)
			if url != tt.want {
				t.Fatalf("url = %q, want %q", url, tt.want)
			}
		})
	}
}

func TestKiroAuthOutputStatus(t *testing.T) {
	if status := kiroAuthOutputStatus("Signed in with Google"); status != kiroAuthSucceeded {
		t.Fatalf("status = %q, want success", status)
	}
	if status := kiroAuthOutputStatus("Logged in successfully"); status != kiroAuthSucceeded {
		t.Fatalf("status = %q, want success", status)
	}
	if status := kiroAuthOutputStatus("device code expired"); status != kiroAuthFailed {
		t.Fatalf("status = %q, want failed", status)
	}
	if status := kiroAuthOutputStatus("still waiting"); status != kiroAuthPending {
		t.Fatalf("status = %q, want pending", status)
	}
}
