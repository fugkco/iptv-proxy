package server

import "testing"

func Test_shouldProxy(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{
			name: ".ts proxies",
			line: "test.ts",
			want: true,
		},
		{
			name: ".m3u proxies",
			line: "test.m3u",
			want: true,
		},
		{
			name: ".m3u8 proxies",
			line: "test.m3u8",
			want: true,
		},
		{
			name: "http://* proxies",
			line: "http://wotm8",
			want: true,
		},
		{
			name: "https://* proxies",
			line: "https://wotm8",
			want: true,
		},
		{
			name: "#EXTINF does not proxy",
			line: "#EXTINF: -1, Something",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldProxy(tt.line); got != tt.want {
				t.Errorf("shouldProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}