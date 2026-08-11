package chat_token

import "testing"

func TestIsIPAllowed(t *testing.T) {
	tests := []struct {
		name      string
		clientIP  string
		whitelist []string
		want      bool
	}{
		{
			name:      "exact match",
			clientIP:  "127.0.0.1",
			whitelist: []string{"127.0.0.1"},
			want:      true,
		},
		{
			name:      "docker subnet match",
			clientIP:  "172.20.0.12",
			whitelist: []string{"172.20.0.0/16"},
			want:      true,
		},
		{
			name:      "outside subnet",
			clientIP:  "172.21.0.12",
			whitelist: []string{"172.20.0.0/16"},
			want:      false,
		},
		{
			name:      "invalid entry does not allow",
			clientIP:  "172.30.0.12",
			whitelist: []string{"not-an-ip"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIPAllowed(tt.clientIP, tt.whitelist); got != tt.want {
				t.Fatalf("isIPAllowed(%q, %v) = %v, want %v", tt.clientIP, tt.whitelist, got, tt.want)
			}
		})
	}
}
