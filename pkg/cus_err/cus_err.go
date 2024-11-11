package cus_err

import (
	"encoding/json"
	"errors"
	"fmt"
	internal "go_micro_service_api/pkg/cus_err/internal/gen"
	"net/http"

	"google.golang.org/grpc/status"
)

// CusError is a custom error type that wraps error code, message, and error sources.
type CusError struct {
	code    CusCode
	msg     string
	data    any
	sources []error
}

// New creates a new CusError.
// Parameters:
//   - code: The CusCode of the error.
//   - msg: The error message.
//   - data: The data associated with the error.
//   - source: The error sources.
//
// Returns:
//   - error: The CusError.
//
// Example:
//
//	err := NewCusError(ErrInvalidInput, "err_msg")
//	err := NewCusError(ErrInvalidInput, "err_msg",err)
//	err := NewCusError(ErrInvalidInput, "err_msg",err1,err2)
func New(code CusCode, msg string, source ...error) *CusError {
	return &CusError{
		code:    code,
		msg:     msg,
		sources: source,
	}
}

// Error returns a string representation of the CusError.
// if there are sources, it will include the sources in the string.
func (e *CusError) Error() string {
	if len(e.sources) > 0 {
		return fmt.Sprintf("cusCode: %v, msg:%s,  sources: %v", e.code, e.msg, e.sources)
	}
	return fmt.Sprintf("cusCode: %v, msg:%s", e.code, e.msg)
}

// HttpCode returns the standard HTTP status code.
func (e *CusError) HttpCode() int {
	// Check self is nil
	if e == nil {
		return http.StatusInternalServerError
	}

	return e.code.HttpCode()
}

// Code returns the cus error code.
func (e *CusError) Code() CusCode {
	return e.code
}

// Message returns the error message.
func (e *CusError) Message() string {
	return e.msg
}

// Unwrap returns the error sources.
// if there are no sources, it will return nil.
func (e *CusError) Unwrap() []error {
	return e.sources
}

// WithData add data to the CusError.
// Parameters:
//   - data: The data to be added to the CusError.
//
// Returns:
//   - error: The CusError.
//
// Example:
//
//	data := map[string]interface{}{"key1": "value1"}
//	err := NewCusError(ErrInvalidInput, "err_msg").WithData(data)
func (e *CusError) WithData(data any) *CusError {
	e.data = data
	return e
}

// Is checks if the target error matches the CusError.
func (e *CusError) Is(target error) bool {
	t, ok := target.(*CusError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// Data returns the data associated with the CusError.
func (e *CusError) Data() any {
	return e.data
}

// WithSource add error sources to the CusError.
func (e *CusError) WithSource(err error) *CusError {
	e.sources = append(e.sources, err)
	return e
}

// FromGrpcErr converts a gRPC error to a CusError.
// Parameters:
//   - err: The gRPC error.
//
// Returns:
//   - error: The CusError.
//   - ok: A boolean indicating if the conversion was successful.
//
// Example:
//
//	cusErr, ok := FromGrpcErr(err)
func FromGrpcErr(err error) (cusErr *CusError, ok bool) {
	st, ok := status.FromError(err)
	if !ok {
		return nil, false
	}

	// Check if the error is our custom CusError
	for _, detail := range st.Details() {
		if proto, ok := detail.(*internal.CusErrorProto); ok {
			cusErr, err := fromProto(proto)
			if err != nil {
				return nil, false
			}
			return cusErr, true
		}
	}

	return nil, false
}

// toProto converts the CusError to a proto message.
func (e *CusError) toProto() (*internal.CusErrorProto, error) {
	dataBytes, err := json.Marshal(e.data)
	if err != nil {
		return nil, err
	}

	sources := make([]string, len(e.sources))
	for i, src := range e.sources {
		if src != nil {
			sources[i] = src.Error()
		}
	}

	return &internal.CusErrorProto{
		Code:    int32(e.code),
		Message: e.msg,
		Data:    dataBytes,
		Source:  sources,
	}, nil
}

// fromProto converts a proto message to a CusError.
func fromProto(proto *internal.CusErrorProto) (*CusError, error) {
	data := make(map[string]interface{})
	if err := json.Unmarshal(proto.Data, &data); err != nil {
		return nil, err
	}

	sources := make([]error, len(proto.Source))
	for i, src := range proto.Source {
		sources[i] = errors.New(src)
	}

	return &CusError{
		code:    CusCode(proto.Code),
		msg:     proto.Message,
		data:    data,
		sources: sources,
	}, nil
}
