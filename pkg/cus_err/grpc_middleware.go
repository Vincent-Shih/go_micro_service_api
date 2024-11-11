package cus_err

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// ErrorInterceptor is a gRPC unary interceptor that converts custom CusError to gRPC error with details
//
// Usage:
//
//	s := grpc.NewServer(
//		grpc.ChainUnaryInterceptor(
//			cuserr.ErrorInterceptor,
//			// Any other interceptors can be added here
//		),
func ErrorInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	// Check if the error is our custom CusError
	if cusErr, ok := err.(*CusError); ok {
		// Convert CusError to gRPC status
		st := status.New(cusErr.Code().GrpcCode(), cusErr.Error())

		// Convert CusError to proto
		proto, err := cusErr.toProto()
		if err != nil {
			// If we can't add the details, just return the status as is
			return nil, st.Err()
		}
		// Add the details to the status
		detailedStatus, err := st.WithDetails(proto)
		if err != nil {
			// If we can't add the details, just return the status as is
			return nil, st.Err()
		}

		return nil, detailedStatus.Err()
	}

	// If it's not a CusError, return the original error
	return nil, err
}

// StreamErrorInterceptor is a gRPC stream interceptor that converts custom CusError to gRPC error with details
//
// Usage:
//
//	s := grpc.NewServer(
//		grpc.StreamInterceptor(
//			cuserr.StreamErrorInterceptor,
//			// Any other interceptors can be added here
//		),
func StreamErrorInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, ss)
	if err == nil {
		return nil
	}

	// Check if the error is our custom CusError
	if cusErr, ok := err.(*CusError); ok {
		// Convert CusError to gRPC status
		st := status.New(cusErr.Code().GrpcCode(), cusErr.Error())

		// Convert CusError to proto
		proto, err := cusErr.toProto()
		if err != nil {
			// If we can't add the details, just return the status as is
			return st.Err()
		}
		// Add the details to the status
		detailedStatus, err := st.WithDetails(proto)
		if err != nil {
			// If we can't add the details, just return the status as is
			return st.Err()
		}

		return detailedStatus.Err()
	}

	// If it's not a CusError, return the original error
	return err
}
