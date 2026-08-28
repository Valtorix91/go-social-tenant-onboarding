package main

import "testing"

func TestSignupRequiresCaptcha(t *testing.T) {
	cases := []struct {
		name, email, token string
		want               bool
	}{
		{"both present", "owner@example.com", "proof", true},
		{"missing email", "", "proof", false},
		{"missing token", "owner@example.com", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validSignup(tc.email, tc.token); got != tc.want {
				t.Fatalf("validSignup() = %v, want %v", got, tc.want)
			}
		})
	}
}
