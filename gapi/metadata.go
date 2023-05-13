package gapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeaderKey = "grpcgateway-user-agent"
	userAgentHeader               = "user-agent"
	xForwardedForHeader           = "x-forwarded-for"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return mtdt
	}
	if userAgents := md.Get(grpcGatewayUserAgentHeaderKey); len(userAgents) > 0 {
		mtdt.UserAgent = userAgents[0]
	} else if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
		mtdt.UserAgent = userAgents[0]
	}
	if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
		mtdt.ClientIP = clientIPs[0]
	} else if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = p.Addr.String()
	}

	return mtdt
}
