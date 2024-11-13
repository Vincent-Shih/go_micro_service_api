package vo

import (
	"context"
	"fmt"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"time"
)

const (
	_clientIdKey = "cid"
	_userIdKey   = "uid"
	_roleIdKey   = "rid"
	_accountKey  = "acc"
	_merchantId  = "mid"
	_issueAt     = "iat"
)

type TokenPayload struct {
	MerchantId int64   // Merchant id is required
	ClientId   int64   // Client id is required
	IssueAt    int64   // Issue at is required
	UserId     *int64  // User if could be nil
	Account    *string // Account could be nil
	RoleId     *int64  // Role id could be nil

}

type TokenPayloadOption func(*TokenPayload)

func WithUserId(userId int64) TokenPayloadOption {
	return func(tp *TokenPayload) {
		tp.UserId = &userId
	}
}

func WithRoleId(roleId int64) TokenPayloadOption {
	return func(tp *TokenPayload) {
		tp.RoleId = &roleId
	}
}

func WithAccount(account string) TokenPayloadOption {
	return func(tp *TokenPayload) {
		tp.Account = &account
	}
}

func NewTokenPayload(merchantId int64, clientId int64, opts ...TokenPayloadOption) TokenPayload {
	tp := TokenPayload{
		MerchantId: merchantId,
		ClientId:   clientId,
		IssueAt:    time.Now().Unix(),
	}

	for _, opt := range opts {
		opt(&tp)
	}

	return tp
}

func ToTokenPayload(ctx context.Context, payload map[string]interface{}) (TokenPayload, *cus_err.CusError) {
	tp := TokenPayload{}

	// Try to get client id from payload
	// the client id is required , if not found return error
	cidVal, exists := payload[_clientIdKey]
	if !exists {
		cusErr := cus_err.New(cus_err.InternalServerError, fmt.Sprintf("Missing %s in token payload", _clientIdKey))
		cus_otel.Error(ctx, cusErr.Error())
		return tp, cusErr
	}
	cid, ok := cidVal.(float64)
	if !ok {
		cusErr := cus_err.New(cus_err.InternalServerError, fmt.Sprintf("Invalid %s in token payload", _clientIdKey))
		cus_otel.Error(ctx, cusErr.Error())
		return tp, cusErr
	}
	tp.ClientId = int64(cid)

	// Try to get issue at from payload
	// the issue at is required , if not found return error
	iatVal, exists := payload[_issueAt]
	if !exists {
		cusErr := cus_err.New(cus_err.InternalServerError, fmt.Sprintf("Missing %s in token payload", _issueAt))
		cus_otel.Error(ctx, cusErr.Error())
		return tp, cusErr
	}
	iat, ok := iatVal.(float64)
	if !ok {
		cusErr := cus_err.New(cus_err.InternalServerError, fmt.Sprintf("Invalid %s in token payload", _issueAt))
		cus_otel.Error(ctx, cusErr.Error())
		return tp, cusErr
	}
	tp.IssueAt = int64(iat)

	// Try to get merchant id from payload
	// the merchant id is required , if not found return error
	midVal, exists := payload[_merchantId]
	if !exists {
		cusErr := cus_err.New(cus_err.InternalServerError, fmt.Sprintf("Missing %s in token payload", _merchantId))
		cus_otel.Error(ctx, cusErr.Error())
		return tp, cusErr
	}
	mid, ok := midVal.(float64)
	if !ok {
		cusErr := cus_err.New(cus_err.InternalServerError, fmt.Sprintf("Invalid %s in token payload", _merchantId))
		cus_otel.Error(ctx, cusErr.Error())
		return tp, cusErr
	}
	tp.MerchantId = int64(mid)

	// Try to get user id and role id from payload
	if uidVal, exists := payload[_userIdKey]; exists {
		if uid, ok := uidVal.(float64); ok {
			uid64 := int64(uid)
			tp.UserId = &uid64
		}
	}

	// Try to get role id from payload
	if ridVal, exists := payload[_roleIdKey]; exists {
		if rid, ok := ridVal.(float64); ok {
			rid64 := int64(rid)
			tp.RoleId = &rid64
		}
	}

	// Try to get account from payload
	if accVal, exists := payload[_accountKey]; exists {
		if acc, ok := accVal.(string); ok {
			tp.Account = &acc
		}
	}

	return tp, nil
}

func (t *TokenPayload) ToMap() map[string]any {
	payload := map[string]any{
		_clientIdKey: t.ClientId,
		_merchantId:  t.MerchantId,
		_issueAt:     t.IssueAt,
	}

	if t.UserId != nil {
		payload[_userIdKey] = *t.UserId
	}

	if t.RoleId != nil {
		payload[_roleIdKey] = *t.RoleId
	}

	if t.Account != nil {
		payload[_accountKey] = *t.Account
	}

	return payload
}
