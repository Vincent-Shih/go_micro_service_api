package cus_err

import (
	"errors"
	internal "go_micro_service_api/pkg/cus_err/internal/gen"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestCusCodeHttpCode tests the HttpCode method of CusCode
func TestCusCodeHttpCode(t *testing.T) {
	tests := []struct {
		name     string
		code     CusCode
		expected int
	}{
		{"OK", OK, http.StatusOK},
		{"BadRequest", AccountError, http.StatusBadRequest},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized},
		{"StatusNotFound", ResponseNotFound, http.StatusNotFound},
		{"InternalServerError", InternalServerError, http.StatusInternalServerError},
		{"NotImplemented", NotImplemented, http.StatusNotImplemented},
		{"InvalidCode", 999_9999, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.code.HttpCode())
		})
	}
}

// TestNewCusError tests the New function for creating CusError
func TestNewCusError(t *testing.T) {
	err := New(AccountError, "test error")
	assert.Equal(t, AccountError, err.Code().Int())
	assert.Equal(t, "test error", err.Message())
	assert.Nil(t, err.Data())
	assert.Empty(t, err.Unwrap())

	sourceErr := errors.New("source error")
	err = New(AccountError, "test error", sourceErr)
	assert.Equal(t, AccountError, err.Code().Int())
	assert.Equal(t, "test error", err.Message())
	assert.Nil(t, err.Data())
	assert.Equal(t, []error{sourceErr}, err.Unwrap())
}

// TestCusErrorError tests the Error method of CusError
func TestCusErrorError(t *testing.T) {
	err := New(AccountError, "test error")
	assert.Equal(t, "kgsCode: 4000000, msg:test error", err.Error())

	sourceErr := errors.New("source error")
	err = New(AccountError, "test error", sourceErr)
	assert.Equal(t, "kgsCode: 4000000, msg:test error,  sources: [source error]", err.Error())
}

// TestCusErrorHttpCode tests the HttpCode method of CusError
func TestCusErrorHttpCode(t *testing.T) {
	err := New(AccountError, "test error")
	assert.Equal(t, http.StatusBadRequest, err.HttpCode())

	var nilErr *CusError
	assert.Equal(t, http.StatusInternalServerError, nilErr.HttpCode())
}

// TestCusErrorWithData tests the WithData method of CusError
func TestCusErrorWithData(t *testing.T) {
	data := map[string]string{"key": "value"}
	err := New(AccountError, "test error").WithData(data)
	assert.Equal(t, data, err.Data())
}

// TestCusErrorIs tests the Is method of CusError
func TestCusErrorIs(t *testing.T) {
	err1 := New(AccountError, "test error")
	err2 := New(AccountError, "another test error")
	err3 := New(Unauthorized, "unauthorized error")

	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
	assert.False(t, err1.Is(errors.New("standard error")))
}

// TestCusErrorWithSource tests the WithSource method of CusError
func TestCusErrorWithSource(t *testing.T) {
	sourceErr := errors.New("source error")
	err := New(AccountError, "test error").WithSource(sourceErr)
	assert.Equal(t, []error{sourceErr}, err.Unwrap())
}

func TestFromGrpcErr(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr *CusError
		wantOk  bool
	}{
		{
			name:    "Non-gRPC error",
			err:     errors.New("regular error"),
			wantErr: nil,
			wantOk:  false,
		},
		{
			name:    "gRPC error without CusErrorProto",
			err:     status.Error(codes.NotFound, "not found"),
			wantErr: nil,
			wantOk:  false,
		},
		{
			name: "gRPC error with CusErrorProto",
			err: func() error {
				st := status.New(codes.Internal, "internal error")
				kgsProto, _ := (&CusError{
					code:    500,
					msg:     "test error",
					data:    map[string]interface{}{"key": "value"},
					sources: []error{errors.New("source error")},
				}).toProto()
				st, _ = st.WithDetails(kgsProto)
				return st.Err()
			}(),
			wantErr: &CusError{
				code:    500,
				msg:     "test error",
				data:    map[string]interface{}{"key": "value"},
				sources: []error{errors.New("source error")},
			},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr, gotOk := FromGrpcErr(tt.err)
			assert.Equal(t, tt.wantOk, gotOk)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.code, gotErr.code)
				assert.Equal(t, tt.wantErr.msg, gotErr.msg)
				assert.Equal(t, tt.wantErr.data, gotErr.data)
				assert.Equal(t, len(tt.wantErr.sources), len(gotErr.sources))
				for i := range tt.wantErr.sources {
					assert.Equal(t, tt.wantErr.sources[i].Error(), gotErr.sources[i].Error())
				}
			} else {
				assert.Nil(t, gotErr)
			}
		})
	}
}

func TestCusErrorToProto(t *testing.T) {
	t.Run("Normal CusError", func(t *testing.T) {
		cusErr := &CusError{
			code:    400,
			msg:     "bad request",
			data:    map[string]interface{}{"reason": "invalid input"},
			sources: []error{errors.New("validation failed")},
		}

		proto, err := cusErr.toProto()
		assert.NoError(t, err)
		assert.NotNil(t, proto)

		assert.Equal(t, int32(400), proto.Code)
		assert.Equal(t, "bad request", proto.Message)
		assert.JSONEq(t, `{"reason":"invalid input"}`, string(proto.Data))
		assert.Equal(t, []string{"validation failed"}, proto.Source)
	})

	t.Run("nil data and sources", func(t *testing.T) {
		cusErr := &CusError{
			code: 500,
			msg:  "internal error",
		}

		proto, err := cusErr.toProto()
		assert.NoError(t, err)
		assert.NotNil(t, proto)

		assert.Equal(t, int32(500), proto.Code)
		assert.Equal(t, "internal error", proto.Message)
		assert.Empty(t, proto.Source)
	})
}

func TestFromProto(t *testing.T) {
	t.Run("Normal CusErrorProto", func(t *testing.T) {
		proto := &internal.CusErrorProto{
			Code:    int32(404),
			Message: "not found",
			Data:    []byte(`{"id":"123"}`),
			Source:  []string{"database error"},
		}

		cusErr, err := fromProto(proto)
		assert.NoError(t, err)
		assert.NotNil(t, cusErr)

		assert.Equal(t, CusCode(404), cusErr.code)
		assert.Equal(t, "not found", cusErr.msg)
		assert.Equal(t, map[string]interface{}{"id": "123"}, cusErr.data)
		assert.Equal(t, 1, len(cusErr.sources))
		assert.Equal(t, "database error", cusErr.sources[0].Error())
	})

	t.Run("Invalid JSON data", func(t *testing.T) {
		proto := &internal.CusErrorProto{
			Code:    int32(404),
			Message: "not found",
			Data:    []byte(`invalid json`),
			Source:  []string{"database error"},
		}

		cusErr, err := fromProto(proto)
		assert.Error(t, err)
		assert.Nil(t, cusErr)
	})

	t.Run("Empty source", func(t *testing.T) {
		proto := &internal.CusErrorProto{
			Code:    int32(404),
			Message: "not found",
			Data:    []byte(`{"id":"123"}`),
		}

		cusErr, err := fromProto(proto)
		assert.NoError(t, err)
		assert.NotNil(t, cusErr)

		assert.Equal(t, CusCode(404), cusErr.code)
		assert.Equal(t, "not found", cusErr.msg)
		assert.Equal(t, map[string]interface{}{"id": "123"}, cusErr.data)
		assert.Empty(t, cusErr.sources)
	})
}
