package node

import (
	"context"

	"github.com/aos-dev/noah/proto"
)

type Portal struct {
	proto.UnimplementedNodeServer
}

func (p Portal) Register(ctx context.Context, request *proto.RegisterRequest) (*proto.RegisterReply, error) {
	panic("implement me")
}

func (p Portal) mustEmbedUnimplementedAgentServer() {
	panic("implement me")
}
