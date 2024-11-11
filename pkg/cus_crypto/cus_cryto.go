package cus_crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"

	"golang.org/x/crypto/bcrypt"
)

type CusCrypto struct{}

// AESKey is a struct that holds the key and IV for AES encryption.
type AESKey struct {
	Key string
	IV  string
}

// New creates a new CusCrypto service.
func New() CusCrypto {
	return CusCrypto{}
}

// GenerateRandomSecret generates a random secret of the given length.
func (k *CusCrypto) GenerateRandomSecret(ctx context.Context, length int) ([]byte, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "failed to generate secret", err)
		cus_otel.Error(ctx, cusErr.Error())
		return []byte{}, cusErr
	}

	return bytes, nil
}

// EncryptAES encrypts the given data using AES-CFB encryption.
func (k *CusCrypto) EncryptAES(ctx context.Context, key AESKey, data string) ([]byte, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	plaintext := []byte(data)

	block, err := aes.NewCipher([]byte(key.Key))
	if err != nil {
		cus_otel.Error(ctx, err.Error())
		return []byte{}, cus_err.New(cus_err.InternalServerError, "failed to create cipher", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := []byte(key.IV)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// DecryptAES256 decrypts the given data using AES-CFB decryption.
func (k *CusCrypto) DecryptAES(ctx context.Context, key AESKey, data []byte) (string, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	block, err := aes.NewCipher([]byte(key.Key))
	if err != nil {
		cus_otel.Error(ctx, err.Error())
		return "", cus_err.New(cus_err.InternalServerError, "failed to create cipher", err)
	}

	iv := []byte(key.IV)
	plaintext := make([]byte, len(data)-aes.BlockSize)
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plaintext, data[aes.BlockSize:])

	return string(plaintext), nil
}

// EncryptAESCBC encrypts the given data using AES-CBC encryption.
func (k *CusCrypto) EncryptAESCBC(ctx context.Context, key string, data string) ([]byte, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	plaintext := []byte(data)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		cus_otel.Error(ctx, err.Error())
		return []byte{}, cus_err.New(cus_err.InternalServerError, "failed to create cipher", err)
	}

	// Use PKCS7 padding
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	for i := 0; i < padding; i++ {
		plaintext = append(plaintext, byte(padding))
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		cus_otel.Error(ctx, err.Error())
		return []byte{}, cus_err.New(cus_err.InternalServerError, "failed to generate IV", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// DecryptAESCBC decrypts the given data using AES-CBC decryption.
func (k *CusCrypto) DecryptAESCBC(ctx context.Context, key string, data []byte) (string, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		cus_otel.Error(ctx, err.Error())
		return "", cus_err.New(cus_err.InternalServerError, "failed to create cipher", err)
	}

	// Check if the data is long enough to contain the IV
	if len(data) < aes.BlockSize {
		cusErr := cus_err.New(cus_err.AccountPasswordError, "ciphertext too short")
		cus_otel.Error(ctx, cusErr.Error())
		return "", cusErr
	}

	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Unpad the plaintext
	padding := int(ciphertext[len(ciphertext)-1])
	plaintext := ciphertext[:len(ciphertext)-padding]

	return string(plaintext), nil
}

// HashMD5 hashes the given data using MD5.
func (k *CusCrypto) HashMD5(ctx context.Context, data string) []byte {
	_, span := cus_otel.StartTrace(ctx)
	defer span.End()

	hasher := md5.New()

	hasher.Write([]byte(data))

	hashInbytes := hasher.Sum(nil)

	return hashInbytes
}

// HashSHA256 hashes the given data using SHA-256.
func (k *CusCrypto) HashSHA256(ctx context.Context, data string) []byte {
	_, span := cus_otel.StartTrace(ctx)
	defer span.End()

	hasher := sha256.New()

	hasher.Write([]byte(data))

	hashInbytes := hasher.Sum(nil)

	return hashInbytes
}

// HashPassword hashes the given password using bcrypt.
func (k *CusCrypto) HashPassword(ctx context.Context, data string) (string, *cus_err.CusError) {
	_, span := cus_otel.StartTrace(ctx)
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "failed to hash bcrypt", err)
		cus_otel.Error(ctx, cusErr.Error())
		return "", cusErr
	}

	return string(hash), nil
}

// CompareHashAndPassword compares the given hash and password using bcrypt.
// Returns true if the password matches the hash.
func (k *CusCrypto) CompareHashAndPassword(ctx context.Context, hash, password string) bool {
	_, span := cus_otel.StartTrace(ctx)
	defer span.End()

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// EncodeHex encodes the given data to a hex string.
func (k *CusCrypto) EncodeHex(ctx context.Context, data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeHex decodes the given hex string to a byte array.
func (k *CusCrypto) DecodeHex(ctx context.Context, data string) ([]byte, *cus_err.CusError) {
	bytes, err := hex.DecodeString(data)
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "failed to decode hex", err)
		cus_otel.Error(ctx, cusErr.Error())
		return []byte{}, cusErr
	}

	return bytes, nil
}

// EncodeBase64 encodes the given data to a base64 string.
func (k *CusCrypto) EncodeBase64(ctx context.Context, data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeBase64 decodes the given base64 string to a byte array.
func (k *CusCrypto) DecodeBase64(ctx context.Context, data string) ([]byte, *cus_err.CusError) {
	bytes, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "failed to decode base64", err)
		cus_otel.Error(ctx, cusErr.Error())
		return []byte{}, cusErr
	}

	return bytes, nil
}
