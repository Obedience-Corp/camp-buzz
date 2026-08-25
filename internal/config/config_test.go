package config

import "testing"

func TestNormalizeRelayURL(t *testing.T) {
	cases := map[string]string{
		"ws://localhost:3000":   "http://localhost:3000",
		"wss://relay.example":   "https://relay.example",
		"http://localhost:3000": "http://localhost:3000",
		"https://relay.example": "https://relay.example",
		"":                      "",
	}
	for in, want := range cases {
		if got := NormalizeRelayURL(in); got != want {
			t.Fatalf("NormalizeRelayURL(%q)=%q want %q", in, got, want)
		}
	}
}
