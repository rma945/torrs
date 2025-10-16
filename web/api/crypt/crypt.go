package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"log/slog"
)

func GetKeyPair() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		slog.Error("Failed to get key pair", "err", err)
	}
	return key
}

func Decrypt(key *rsa.PrivateKey, data string) string {
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err == nil {
		textBytes, err := rsa.DecryptPKCS1v15(rand.Reader, key, bytes)
		if err == nil {
			return string(textBytes)
		}
	}
	slog.Error("Failed to decrypt data", "err", err)
	return ""
}

func Encrypt(key *rsa.PublicKey, data string) string {
	cryptBytes, err := rsa.EncryptPKCS1v15(rand.Reader, key, []byte(data))
	if err == nil {
		return base64.StdEncoding.EncodeToString(cryptBytes)
	}
	slog.Error("Failed to encrypt data", "err", err)
	return ""
}
