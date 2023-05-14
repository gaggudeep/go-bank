package gapi

import (
	"context"
	cnst "github.com/gaggudeep/bank_go/const"
	"github.com/gaggudeep/bank_go/const/enum"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

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
		Str(cnst.LogProtocol, enum.ProtocolGRPC.String()).
		Str(cnst.LogMethod, info.FullMethod).
		Int(cnst.LogStatusCode, int(statusCode)).
		Str(cnst.LogStatusText, statusCode.String()).
		Dur(cnst.LogDuration, duration).
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
				Bytes(cnst.LogBody, rec.Body)
		} else {
			logger = log.Info()
		}

		logger.
			Str(cnst.LogProtocol, enum.ProtocolHTTP.String()).
			Str(cnst.LogMethod, req.Method).
			Str(cnst.LogPath, req.RequestURI).
			Int(cnst.LogStatusCode, rec.StatusCode).
			Str(cnst.LogStatusText, http.StatusText(rec.StatusCode)).
			Dur(cnst.LogDuration, duration).
			Msg("received an HTTP request")
	})
}
