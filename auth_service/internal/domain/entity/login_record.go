package entity

import (
	"context"
	"time"
)

// LoginRecord represents a record of a user's login.
type LoginRecord struct {
	Id          int64
	Browser     string
	BrowserVer  string
	Ip          string
	Os          string
	Platform    string
	Country     string
	CountryCode string
	City        string
	Asp         string
	IsMobile    bool
	IsSuccess   bool
	CreateAt    time.Time
	ErrMessage  string
}

func NewLoginRecord(ctx context.Context, userId int64, isSuccess bool) *LoginRecord {
	r := &LoginRecord{}

	if browser, ok := ctx.Value("browser").(string); ok {
		r.Browser = browser
	}

	if ip, ok := ctx.Value("ip").(string); ok {
		r.Ip = ip
	}

	if os, ok := ctx.Value("os").(string); ok {
		r.Os = os
	}

	if country, ok := ctx.Value("country").(string); ok {
		r.Country = country
	}

	if city, ok := ctx.Value("city").(string); ok {
		r.City = city
	}

	r.IsSuccess = isSuccess

	return r
}
