package agent

import (
	"context"
	"github.com/aos-dev/noah/proto"
)

type Server struct {
	proto.UnimplementedAgentServer
}

func (s Server) GetEndpoint(ctx context.Context, request *proto.GetEndpointRequest) (*proto.GetEndpointReply, error) {
	panic("implement me")
}

func (s Server) mustEmbedUnimplementedAgentServer() {
	panic("implement me")
}
