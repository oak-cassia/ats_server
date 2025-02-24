package service

import "testing"

func TestGenerateToken(t *testing.T) {
	token := generateToken()
	if len(token) != 44 {
		t.Errorf("token size is not 32, got: %d", len(token))
	}
}
