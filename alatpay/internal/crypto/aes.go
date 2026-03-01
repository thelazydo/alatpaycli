package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// PKCS5Padding pads plaintext to block size
// (Technically PKCS7 for 16-byte blocks, which AlatPay expects)
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS5UnPadding unpads plaintext
func PKCS5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("empty data")
	}
	unpadding := int(origData[length-1])
	if unpadding > length {
		return nil, errors.New("unpad error")
	}
	return origData[:(length - unpadding)], nil
}

// Encrypt encrypts a plaintext string using AES CBC with PKCS5 padding
// and returns the base64 encoded ciphertext string.
func Encrypt(plaintext string, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	pt := PKCS5Padding([]byte(plaintext), block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))

	ciphertext := make([]byte, len(pt))
	blockMode.CryptBlocks(ciphertext, pt)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64 encoded string using AES CBC and PKCS5 unpadding.
func Decrypt(b64Ciphertext string, key, iv string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(b64Ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(ciphertext) < block.BlockSize() {
		return "", errors.New("ciphertext too short")
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	plaintext := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := PKCS5UnPadding(plaintext)
	if err != nil {
		return "", err
	}

	return string(unpadded), nil
}
