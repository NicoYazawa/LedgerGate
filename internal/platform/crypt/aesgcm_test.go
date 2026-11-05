package crypt

import (
	"bytes"
	"errors"
	"testing"
)

func TestAESGCMRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	ciphertext, err := NewAESGCM(key)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("provider-api-key")

	encrypted, err := ciphertext.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if bytes.Contains(encrypted, plaintext) {
		t.Fatal("ciphertext contains plaintext")
	}
	decrypted, err := ciphertext.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}

	second, err := ciphertext.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(encrypted, second) {
		t.Error("Encrypt() reused the nonce")
	}
}

func TestAESGCMRejectsWrongKeyAndTampering(t *testing.T) {
	ciphertext, err := NewAESGCM(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := ciphertext.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	wrongKey, err := NewAESGCM(bytes.Repeat([]byte{2}, 32))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wrongKey.Decrypt(encrypted); err == nil {
		t.Error("Decrypt() wrong key error = nil")
	}

	tampered := append([]byte(nil), encrypted...)
	tampered[len(tampered)-1] ^= 0xff
	if _, err := ciphertext.Decrypt(tampered); err == nil {
		t.Error("Decrypt() tampered ciphertext error = nil")
	}

	short := encrypted[:12]
	if _, err := ciphertext.Decrypt(short); err == nil {
		t.Error("Decrypt() short input error = nil")
	}
}

func TestAESGCMRejectsInvalidKeyLength(t *testing.T) {
	for _, size := range []int{0, 16, 24, 31, 33} {
		_, err := NewAESGCM(make([]byte, size))
		if err == nil {
			t.Errorf("NewAESGCM(key length %d) error = nil", size)
		}
	}
}

func TestEmptyPlaintextRoundTrip(t *testing.T) {
	ciphertext, err := NewAESGCM(bytes.Repeat([]byte{3}, 32))
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := ciphertext.Encrypt(nil)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := ciphertext.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if len(decrypted) != 0 {
		t.Errorf("Decrypt() length = %d, want 0", len(decrypted))
	}
}

func TestDecryptErrorIsStable(t *testing.T) {
	ciphertext, err := NewAESGCM(bytes.Repeat([]byte{4}, 32))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ciphertext.Decrypt(nil); !errors.Is(err, ErrInvalidCiphertext) {
		t.Errorf("Decrypt() error = %v, want ErrInvalidCiphertext", err)
	}
}
