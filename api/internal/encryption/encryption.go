package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

func deriveKeyAndIV(key, salt []byte) (k32, iv16 []byte) {
	var buf []byte
	prev := key
	for len(buf) < 48 {
		d := md5.Sum(append(prev, salt...))
		buf = append(buf, d[:]...)
		prev = d[:]
	}
	return buf[:32], buf[32:48]
}

func pkcs7Unpad(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, errors.New("empty plaintext")
	}
	pad := int(b[len(b)-1])
	if pad == 0 || pad > aes.BlockSize {
		return nil, errors.New("bad padding size")
	}
	for i := len(b) - pad; i < len(b); i++ {
		if b[i] != byte(pad) {
			return nil, errors.New("bad padding")
		}
	}
	return b[:len(b)-pad], nil
}

func DecryptData(cipherB64, sharedKeyHex string) (string, error) {

	keyBytes, err := hex.DecodeString(sharedKeyHex)
	if err != nil {
		return "", fmt.Errorf("hex decode shared key: %w", err)
	}
	if len(keyBytes) != 32 {
		return "", fmt.Errorf("shared key must be 32 bytes, got %d", len(keyBytes))
	}

	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	if len(raw) < 16 || string(raw[:8]) != "Salted__" {
		return "", errors.New("missing Salted__ prefix")
	}

	salt, ciphertext := raw[8:16], raw[16:]
	key, iv := deriveKeyAndIV(keyBytes, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext size must be multiple of 16")
	}

	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)

	plain, err = pkcs7Unpad(plain)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
