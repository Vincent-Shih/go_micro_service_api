package req_analyzer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserAgentInfo(t *testing.T) {
	r := NewReqAnalyzer()
	t.Run("Test Normal case", func(t *testing.T) {
		str := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
		expected := UserAgentInfo{
			IsMobile:   false,
			IsBot:      false,
			Platform:   "Windows",
			OS:         "Windows 10",
			Browser:    "Chrome",
			BrowserVer: "58.0.3029.110",
		}
		got := r.GetUserAgentInfo(context.Background(), str)
		assert.Equal(t, expected, got)
	})

	t.Run("Test empty str", func(t *testing.T) {
		str := ""
		expected := UserAgentInfo{}
		got := r.GetUserAgentInfo(context.Background(), str)
		assert.Equal(t, expected, got)
	})

	t.Run("Test bot user agent", func(t *testing.T) {
		str := "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
		expected := UserAgentInfo{
			IsMobile:   false,
			IsBot:      true,
			Platform:   "",
			OS:         "",
			Browser:    "Googlebot",
			BrowserVer: "2.1",
		}
		got := r.GetUserAgentInfo(context.Background(), str)
		assert.Equal(t, expected, got)
	})

	t.Run("Test mobile user agent", func(t *testing.T) {
		str := "Mozilla/5.0 (Linux; Android 10; SM-G960U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.93 Mobile Safari/537.36"
		expected := UserAgentInfo{
			IsMobile:   true,
			IsBot:      false,
			Platform:   "Linux",
			OS:         "Android 10",
			Browser:    "Chrome",
			BrowserVer: "88.0.4324.93",
		}
		got := r.GetUserAgentInfo(context.Background(), str)
		assert.Equal(t, expected, got)
	})
}

func TestGetIPInfo(t *testing.T) {
	r := NewReqAnalyzer()
	t.Run("Test Normal case", func(t *testing.T) {
		ip := "140.112.41.24"
		ipInfo := r.GetIpInfo(context.Background(), ip)
		assert.Equal(t, ip, ipInfo.Ip)
		assert.Equal(t, "Taiwan", ipInfo.Country)
		assert.Equal(t, "Taipei", ipInfo.City)
		assert.NotEmpty(t, ipInfo.Asp)
	})

	t.Run("Test empty ip", func(t *testing.T) {
		ip := ""
		ipInfo := r.GetIpInfo(context.Background(), ip)
		assert.Equal(t, ip, ipInfo.Ip)
		assert.Empty(t, ipInfo.Country)
		assert.Empty(t, ipInfo.City)
		assert.Empty(t, ipInfo.Asp)
	})

	t.Run("Test invalid ip", func(t *testing.T) {
		ip := "ejwdsdsadadad=="
		ipInfo := r.GetIpInfo(context.Background(), ip)
		assert.Equal(t, ip, ipInfo.Ip)
		assert.Empty(t, ipInfo.Country)
		assert.Empty(t, ipInfo.City)
		assert.Empty(t, ipInfo.Asp)
	})

	t.Run("Test localhost", func(t *testing.T) {
		ip := "localhost"
		ipInfo := r.GetIpInfo(context.Background(), ip)
		assert.Equal(t, ip, ipInfo.Ip)
		assert.Empty(t, ipInfo.Country)
		assert.Empty(t, ipInfo.City)
		assert.Empty(t, ipInfo.Asp)
	})

	t.Run("Test internal ip", func(t *testing.T) {
		ip := "192.168.0.5"
		ipInfo := r.GetIpInfo(context.Background(), ip)
		assert.Equal(t, ip, ipInfo.Ip)
		assert.Empty(t, ipInfo.Country)
		assert.Empty(t, ipInfo.City)
		assert.Empty(t, ipInfo.Asp)
	})
}
