package gapi

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

const (
	ProtocolType = "protocol"
	Method       = "method"
	Path         = "path"
	StatusCode   = "status_code"
	StatusText   = "status_text"
	Duration     = "duration"
	Body         = "body"
)

type Protocol int8

const (
	ProtocolHTTP Protocol = iota
	ProtocolGRPC
)

func (p Protocol) String() string {
	switch p {
	case ProtocolHTTP:
		return "http"
	case ProtocolGRPC:
		return "gRPC"
	}
	panic("unknown protocol")
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func GRPCLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	startTime := time.Now()
	res, err := handler(ctx, req)
	duration := time.Since(startTime)

	var statusCode codes.Code
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	} else {
		statusCode = codes.Unknown
	}

	var logger *zerolog.Event
	if err != nil {
		logger = log.Error()
	} else {
		logger = log.Info()
	}

	logger.
		Str(ProtocolType, ProtocolGRPC.String()).
		Str(Method, info.FullMethod).
		Int(StatusCode, int(statusCode)).
		Str(StatusText, statusCode.String()).
		Dur(Duration, duration).
		Msg("received a gRPC request")

	return res, err
}

func HTTPLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		var logger *zerolog.Event
		if rec.StatusCode != http.StatusOK {
			logger = log.Error().
				Bytes(Body, rec.Body)
		} else {
			logger = log.Info()
		}

		logger.
			Str(ProtocolType, ProtocolHTTP.String()).
			Str(Method, req.Method).
			Str(Path, req.RequestURI).
			Int(StatusCode, rec.StatusCode).
			Str(StatusText, http.StatusText(rec.StatusCode)).
			Dur(Duration, duration).
			Msg("received an HTTP request")
	})
}
