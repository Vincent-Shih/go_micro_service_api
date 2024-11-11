package cus_crypto

import (
	"context"
	"encoding/hex"
	"go_micro_service_api/pkg/cus_err"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCusCrypto_GenerateRandomSecret(t *testing.T) {
	k := New()
	ctx := context.Background()

	secret, err := k.GenerateRandomSecret(ctx, 32)
	assert.Nil(t, err)
	assert.Equal(t, 32, len(secret))

	secret2, err := k.GenerateRandomSecret(ctx, 32)
	assert.Nil(t, err)
	assert.NotEqual(t, secret, secret2)
}

func TestCusCrypto_EncryptDecryptAES(t *testing.T) {
	k := New()
	ctx := context.Background()
	key := AESKey{
		Key: "0123456789abcdef",
		IV:  "fedcba9876543210",
	}
	plaintext := "Hello, World!"

	ciphertext, err := k.EncryptAES(ctx, key, plaintext)
	assert.Nil(t, err)
	assert.NotEqual(t, plaintext, string(ciphertext))

	decrypted, err := k.DecryptAES(ctx, key, ciphertext)
	assert.Nil(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestCusCrypto_EncryptDecryptAESCBC(t *testing.T) {
	k := New()
	ctx := context.Background()
	key := "0123456789abcdef0123456789abcdef" // 32 bytes for AES-256
	plaintext := "Hello, World!"

	ciphertext, err := k.EncryptAESCBC(ctx, key, plaintext)
	assert.Nil(t, err)
	assert.NotEqual(t, plaintext, string(ciphertext))

	decrypted, err := k.DecryptAESCBC(ctx, key, ciphertext)
	assert.Nil(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestCusCrypto_HashMD5(t *testing.T) {
	k := New()
	ctx := context.Background()
	data := "Hello, World!"
	expected := "65a8e27d8879283831b664bd8b7f0ad4"

	hash := k.HashMD5(ctx, data)
	assert.Equal(t, expected, hex.EncodeToString(hash))
}

func TestCusCrypto_HashSHA256(t *testing.T) {
	k := New()
	ctx := context.Background()
	data := "Hello, World!"
	expected := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	hash := k.HashSHA256(ctx, data)
	assert.Equal(t, expected, hex.EncodeToString(hash))
}

func TestCusCrypto_EncodeDecodeHex(t *testing.T) {
	k := New()
	ctx := context.Background()
	data := []byte("Hello, World!")

	encoded := k.EncodeHex(ctx, data)
	decoded, err := k.DecodeHex(ctx, encoded)
	assert.Nil(t, err)
	assert.Equal(t, data, decoded)
}

func TestCusCrypto_EncodeDecodeBase64(t *testing.T) {
	k := New()
	ctx := context.Background()
	data := []byte("Hello, World!")

	encoded := k.EncodeBase64(ctx, data)
	decoded, err := k.DecodeBase64(ctx, encoded)
	assert.Nil(t, err)
	assert.Equal(t, data, decoded)
}

func TestCusCrypto_DecodeHex_Error(t *testing.T) {
	k := New()
	ctx := context.Background()
	invalidHex := "invalid hex"

	_, err := k.DecodeHex(ctx, invalidHex)
	assert.NotNil(t, err)
	assert.Equal(t, cus_err.InternalServerError, err.Code().Int())
}

func TestCusCrypto_DecodeBase64_Error(t *testing.T) {
	k := New()
	ctx := context.Background()
	invalidBase64 := "invalid base64!"

	_, err := k.DecodeBase64(ctx, invalidBase64)
	assert.NotNil(t, err)
	assert.Equal(t, cus_err.InternalServerError, err.Code().Int())
}

func TestHashPassword(t *testing.T) {
	k := New()
	ctx := context.Background()
	data := "Hello, World!"

	hash, err := k.HashPassword(ctx, data)
	assert.Nil(t, err)
	assert.NotEqual(t, data, string(hash))

	match := k.CompareHashAndPassword(ctx, hash, data)
	assert.True(t, match)
}
