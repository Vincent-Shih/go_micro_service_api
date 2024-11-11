package vo

import (
	"context"
	"crypto/md5"
	"fmt"
	"go_micro_service_api/pkg/cus_otel"
	"time"

	"math/rand"
)

const (
	verificationKey = "verification:%s"
)

type VerificationSession struct {
	Type   string `json:"Type"`
	Prefix string `json:"Prefix"`
	Code   string `json:"Code"`
	Token  string `json:"Token"`
}

func NewVerificationSession(event, prefix, code, token string) *VerificationSession {
	return &VerificationSession{
		Type:   event,
		Prefix: prefix,
		Code:   code,
		Token:  token,
	}
}

func (vs *VerificationSession) GetCodeRedisKey() string {
	return fmt.Sprintf(verificationKey, vs.Token)
}

func (vs *VerificationSession) GetErrorCountRedisKey() string {
	return fmt.Sprintf("%s:%s", vs.GetCodeRedisKey(), "errorCount")
}

func (vs *VerificationSession) GetLockRedisKey() string {
	return fmt.Sprintf("%s:%s", vs.GetCodeRedisKey(), "lock")
}

func (vs *VerificationSession) GetNotifyLockRedisKey() string {
	return fmt.Sprintf("%s:%s", vs.GetCodeRedisKey(), "notifyLock")
}

func (vs *VerificationSession) GetVerificationCode() string {
	return fmt.Sprintf("%s-%s", vs.Prefix, vs.Code)
}

func (vs *VerificationSession) NextCode() *VerificationSession {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	vs.Prefix = generatePrefix(r, 3)
	vs.Code = generateCode(r, 6)

	return vs
}

func (vs *VerificationSession) NextToken(ctx context.Context, data ...string) *VerificationSession {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	vs.Token = generateToken(ctx, r, data...)

	return vs
}

func (vs *VerificationSession) Verify(code string, token string) bool {
	return vs.GetVerificationCode() == code && vs.Token == token
}

func generatePrefix(r *rand.Rand, length int) string {
	prefix := make([]byte, length)
	for i := range prefix {
		// 'A' is offset of 65 in ASCII table
		// directly generate byte, instead of sampling index for strings
		prefix[i] = 'A' + byte(r.Intn(26))
	}

	return string(prefix)
}

func generateCode(r *rand.Rand, length int) string {
	code := make([]byte, length)
	for i := range code {
		// '0' is offset of 48 in ASCII table
		code[i] = '0' + byte(r.Intn(10))
	}

	return string(code)
}

func generateToken(ctx context.Context, r *rand.Rand, data ...string) string {
	hasher := md5.New()
	for _, d := range data {
		_, err := hasher.Write([]byte(d))
		if err != nil {
			cus_otel.Warn(ctx, "failed to create token", cus_otel.NewField("error", err.Error()))
			return generateCode(r, 12)
		}
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
