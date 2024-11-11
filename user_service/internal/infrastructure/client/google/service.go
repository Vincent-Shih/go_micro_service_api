package google

import (
	"context"
	"go_micro_service_api/pkg/cus_otel"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

const logKey = "google callback error"
const URL = "https://www.googleapis.com"
const PathGetMe = "/oauth2/v1/tokeninfo"

type Service struct {
	url    *url.URL
	client *resty.Client
}

func NewService(client *resty.Client) *Service {
	serviceUrl, _ := url.Parse(URL)
	return &Service{
		url:    serviceUrl,
		client: client,
	}
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type GetMeResponse struct {
	ID    string `json:"user_id"`
	Email string `json:"email"`
}

// GetMe returns user info
// Parameters:
//   - ctx: context
//   - accessToken: google access token
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
		SetQueryParam("id_token", accessToken). // google uses id_token instead of access_token in old rust codebase
		SetResult(result).
		EnableTrace().
		Get(url)

	cus_otel.TraceRestyResponse(ctx, "google client trace info", url, resp)

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
