package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey        = errors.New("AES-256-GCM requires a 32-byte key")
	ErrInvalidCiphertext = errors.New("invalid AES-GCM ciphertext")
)

type Cipher struct {
	aead cipher.AEAD
}

func NewAESGCM(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

func (ciphertext *Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, ciphertext.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate AES-GCM nonce: %w", err)
	}
	return ciphertext.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (ciphertext *Cipher) Decrypt(encrypted []byte) ([]byte, error) {
	nonceSize := ciphertext.aead.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, ErrInvalidCiphertext
	}
	plaintext, err := ciphertext.aead.Open(nil, encrypted[:nonceSize], encrypted[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("%w: authentication failed", ErrInvalidCiphertext)
	}
	return plaintext, nil
}
