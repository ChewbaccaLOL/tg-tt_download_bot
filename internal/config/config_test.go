package config

import "testing"

func TestIsSupportedURL(t *testing.T) {
	allowed := []string{"tiktok.com", "vm.tiktok.com", "vt.tiktok.com"}

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "tiktok canonical", raw: "https://www.tiktok.com/@user/video/123", want: true},
		{name: "tiktok short", raw: "https://vm.tiktok.com/ZMabc/", want: true},
		{name: "nested supported host", raw: "https://m.tiktok.com/v/123", want: true},
		{name: "unsupported host", raw: "https://example.com/video/123", want: false},
		{name: "not a url", raw: "hello", want: false},
		{name: "unsupported scheme", raw: "ftp://tiktok.com/video/123", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedURL(tt.raw, allowed); got != tt.want {
				t.Fatalf("IsSupportedURL(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
