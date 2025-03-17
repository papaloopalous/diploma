package encryption

import (
	errlist "api/internal/errList"
	"crypto/aes"
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

func GetEncryptionKey() ([]byte, error) {
	key := os.Getenv("ENCRYPTION_KEY")
	decodedKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(key)))
	if err != nil {
		return nil, err
	}
	return decodedKey, nil
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New(errlist.ErrDecrEmpty)
	}

	padding := int(data[length-1])
	if padding == 0 || padding > aes.BlockSize || padding > length {
		return nil, errors.New(errlist.ErrDecrPaddingSize)
	}

	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, errors.New(errlist.ErrDecrPaddindByte)
		}
	}

	return data[:length-padding], nil
}

func DecryptData(encryptedData string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return "", errors.New(errlist.ErrDecrCipher)
	}

	decrypted := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += block.BlockSize() {
		block.Decrypt(decrypted[i:i+block.BlockSize()], ciphertext[i:i+block.BlockSize()])
	}

	decrypted, err = pkcs7Unpad(decrypted)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
