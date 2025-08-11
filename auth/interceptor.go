package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor creates a gRPC unary interceptor for authentication
func UnaryServerInterceptor(requiredToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication if no token is configured
		if requiredToken == "" {
			return handler(ctx, req)
		}

		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header values
		authValues := md.Get("authorization")
		if len(authValues) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Check authorization token
		authHeader := authValues[0]
		if !isValidAuth(authHeader, requiredToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		// Continue with the request
		return handler(ctx, req)
	}
}

// StreamServerInterceptor creates a gRPC stream interceptor for authentication
func StreamServerInterceptor(requiredToken string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip authentication if no token is configured
		if requiredToken == "" {
			return handler(srv, ss)
		}

		// Extract metadata from stream context
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header values
		authValues := md.Get("authorization")
		if len(authValues) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Check authorization token
		authHeader := authValues[0]
		if !isValidAuth(authHeader, requiredToken) {
			return status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		// Continue with the stream
		return handler(srv, ss)
	}
}

// isValidAuth validates the authorization header against the required token
func isValidAuth(authHeader, requiredToken string) bool {
	// Direct comparison for raw authorization values
	return authHeader == requiredToken
}

// ClientInterceptor creates a gRPC client interceptor that adds authorization header
func ClientInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Skip if no token configured
		if token == "" {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Add authorization header to metadata (use token as-is)
		md := metadata.New(map[string]string{
			"authorization": token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Make the call with the modified context
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
