package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HTTPMiddleware returns a gin middleware that adds tracing and metrics
func HTTPMiddleware(t *Telemetry, lgr *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start span
		spanName := c.Request.Method + " " + c.FullPath()
		ctx, span := t.StartSpan(c.Request.Context(), spanName,
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.path", c.FullPath()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// Store span in context
		c.Request = c.Request.WithContext(ctx)

		// Get start time for duration calculation
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.Int64("http.duration_ns", duration.Nanoseconds()),
		)

		// Set span status based on HTTP status code
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, http.StatusText(c.Writer.Status()))
			if c.Errors.Last() != nil {
				span.RecordError(c.Errors.Last().Err)
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Log the request with trace information
		lgr.WithTrace(span).LogHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			duration,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"response_size", c.Writer.Size(),
		)
	}
}

// GRPCUnaryServerInterceptor returns a gRPC interceptor that adds tracing and metrics
func GRPCUnaryServerInterceptor(t *Telemetry, lgr *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Extract trace context
		ctx = t.Extract(ctx, MetadataTextMapCarrier(md))

		// Start span
		ctx, span := t.StartSpan(ctx, info.FullMethod,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
			),
		)
		defer span.End()

		// Get start time for duration calculation
		start := time.Now()

		// Process request
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Add response attributes
		if err != nil {
			st, _ := status.FromError(err)
			span.SetAttributes(attribute.String("rpc.status", st.Code().String()))
			span.SetStatus(codes.Error, st.Message())
			span.RecordError(err)
		} else {
			span.SetAttributes(attribute.String("rpc.status", "OK"))
			span.SetStatus(codes.Ok, "")
		}

		// Log the request with trace information
		lgr.WithTrace(span).LogGRPCRequest(
			info.FullMethod,
			duration,
			err,
		)

		return resp, err
	}
}

// GRPCStreamServerInterceptor returns a gRPC stream interceptor that adds tracing and metrics
func GRPCStreamServerInterceptor(t *Telemetry, lgr *logger.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Extract metadata
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Extract trace context
		ctx = t.Extract(ctx, MetadataTextMapCarrier(md))

		// Start span
		ctx, span := t.StartSpan(ctx, info.FullMethod,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
				attribute.Bool("rpc.stream", true),
			),
		)
		defer span.End()

		// Get start time for duration calculation
		start := time.Now()

		// Create wrapped stream that propagates the context
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Process request
		err := handler(srv, wrappedStream)

		// Calculate duration
		duration := time.Since(start)

		// Add response attributes
		if err != nil {
			st, _ := status.FromError(err)
			span.SetAttributes(attribute.String("rpc.status", st.Code().String()))
			span.SetStatus(codes.Error, st.Message())
			span.RecordError(err)
		} else {
			span.SetAttributes(attribute.String("rpc.status", "OK"))
			span.SetStatus(codes.Ok, "")
		}

		// Log the request with trace information
		lgr.WithTrace(span).LogGRPCRequest(
			info.FullMethod,
			duration,
			err,
		)

		return err
	}
}

// MetadataTextMapCarrier adapts gRPC metadata to TextMapCarrier
type MetadataTextMapCarrier metadata.MD

// Get returns the value associated with the passed key.
func (m MetadataTextMapCarrier) Get(key string) string {
	values := metadata.MD(m).Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set stores the key-value pair.
func (m MetadataTextMapCarrier) Set(key string, value string) {
	metadata.MD(m).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (m MetadataTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// wrappedServerStream wraps a grpc.ServerStream to propagate context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
