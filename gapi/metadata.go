package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader        = "x-forwarded-for"
	UserAgentHeader            = "User-Agent"
)

type MetaData struct {
	UserAgent string
	ClientIp  string
}

func (server *Server) extractMetaData(ctx context.Context) *MetaData {
	md := &MetaData{}
	if mdValue, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := mdValue.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			md.UserAgent = userAgents[0]
		}
		if userAgents := mdValue.Get(UserAgentHeader); len(userAgents) > 0 {
			md.UserAgent = userAgents[0]
		}
		if clientIPs := mdValue.Get(xForwardedForHeader); len(clientIPs) > 0 {
			md.ClientIp = clientIPs[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		md.ClientIp = p.Addr.String()
	}
	return md
}
