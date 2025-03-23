package auth

import (
	"bytes"
	"testing"
)

func TestEmbed(t *testing.T) {
	expected := []byte("-----BEGIN PUBLIC KEY-----")
	if !bytes.Contains(rawPublicKey, expected) {
		t.Errorf("expected %s, but got %s", expected, rawPublicKey)
	}
	expected = []byte("-----BEGIN PRIVATE KEY-----")
	if !bytes.Contains(rawPrivateKey, expected) {
		t.Errorf("expected %s, but got %s", expected, rawPrivateKey)
	}
}
