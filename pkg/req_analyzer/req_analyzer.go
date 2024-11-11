package req_analyzer

import (
	"context"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/req_analyzer/nordvpn"

	"github.com/go-resty/resty/v2"
	"github.com/mssola/useragent"
)

// ReqAnalyzer is a service to analyze request
type ReqAnalyzer interface {
	GetUserAgentInfo(ctx context.Context, userAgent string) UserAgentInfo
	GetIpInfo(ctx context.Context, ip string) IPInfo
}

// reqAnalyzerImpl is an implementation of ReqAnalyzer
type reqAnalyzerImpl struct {
	client *resty.Client
}

// NewReqAnalyzer creates a new ReqAnalyzer
func NewReqAnalyzer() ReqAnalyzer {
	return &reqAnalyzerImpl{
		client: resty.New(),
	}
}

type UserAgentInfo struct {
	IsMobile   bool
	IsBot      bool
	Platform   string
	OS         string
	Browser    string
	BrowserVer string
}

type IPInfo struct {
	Ip          string
	Country     string
	City        string
	Asp         string
	CountryCode string
}

// GetUserAgentInfo extracts and returns detailed information about the user agent string provided.
// It starts a trace for monitoring purposes and ensures the span is ended before returning.
//
// Parameters:
//   - ctx: The context for the request, used for tracing.
//   - userAgent: The user agent string to be analyzed.
//
// Returns:
//   - UserAgentInfo: A struct containing information about the user agent, including whether it is a mobile device,
//     whether it is a bot, the platform, the operating system, the browser, and the browser version.
func (r *reqAnalyzerImpl) GetUserAgentInfo(ctx context.Context, userAgent string) UserAgentInfo {
	_, span := cus_otel.StartTrace(ctx)
	defer span.End()

	if userAgent == "" {
		return UserAgentInfo{}
	}
	ua := useragent.New(userAgent)
	browser, browserVer := ua.Browser()
	return UserAgentInfo{
		IsMobile:   ua.Mobile(),
		IsBot:      ua.Bot(),
		Platform:   ua.Platform(),
		OS:         ua.OS(),
		Browser:    browser,
		BrowserVer: browserVer,
	}
}

// GetIpInfo retrieves information about the given IP address, including its
// country, city, ISP, and country code. It uses the nordvpn service to fetch
// the location data.
//
// Parameters:
//   - ctx: The context for the request, which can be used to control timeouts and cancellations.
//   - ip: The IP address to retrieve information for.
//
// Returns:
//
//	An IPInfo struct containing the IP address, country, city, ISP, and country code.
//	If an error occurs during the retrieval, an empty IPInfo struct is returned.
func (r reqAnalyzerImpl) GetIpInfo(ctx context.Context, ip string) IPInfo {
	res, err := nordvpn.NewService(r.client).GetLocation(ctx, ip)
	if err != nil {
		return IPInfo{}
	}

	return IPInfo{
		Ip:          res.IP,
		Country:     res.Country,
		City:        res.City,
		Asp:         res.ISP,
		CountryCode: res.CountryCode,
	}
}
