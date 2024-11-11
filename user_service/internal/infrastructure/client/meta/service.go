package meta

import (
	"context"
	"go_micro_service_api/pkg/cus_otel"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

const logKey = "meta callback error"
const URL = "https://graph.facebook.com"
const PathGetMe = "/me"

type Service struct {
	url    *url.URL
	client *resty.Client
}

func NewService(client *resty.Client) *Service {
	serviceURL, _ := url.Parse(URL)
	return &Service{
		url:    serviceURL,
		client: client,
	}
}

type ErrorResponse struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FbtraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

type GetMeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// GetMe returns user info
// Parameters:
//   - ctx: context
//   - accessToken: meta access token
//
// Returns:
//   - *GetMeResponse: response body from PathGetMe
//   - error: error
func (s *Service) GetMe(ctx context.Context, accessToken string) (*GetMeResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	url := s.url.JoinPath(PathGetMe).String()
	result := &GetMeResponse{}
	resp, err := s.client.R().
		SetHeader("Accept", "application/json").
		SetQueryParam("access_token", accessToken).
		SetQueryParam("fields", "id,email").
		SetResult(result).
		EnableTrace().
		Get(url)

	cus_otel.TraceRestyResponse(ctx, "meta client trace info", url, resp)

	if resp.StatusCode() != http.StatusOK {
		cus_otel.Error(ctx, logKey, cus_otel.NewField("response_body", resp.String()))
	}

	if err != nil {
		cus_otel.Error(ctx, logKey, cus_otel.NewField("error", err))
		return result, err
	}

	if result.ID == "" {
		cus_otel.Error(ctx, logKey, cus_otel.NewField("error", "empty id"))
		return result, err
	}

	return result, nil
}
