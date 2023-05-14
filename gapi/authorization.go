package gapi

import (
	"context"
	"fmt"
	"github.com/gaggudeep/bank_go/token"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	authorizationHeaderKey    = "authorization"
	authorizationSchemeBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	values := md.Get(authorizationHeaderKey)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authScheme := strings.ToLower(fields[0])
	if authScheme != authorizationSchemeBearer {
		return nil, fmt.Errorf("invalid authorization scheme: %s", authScheme)
	}

	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	return payload, nil
}
