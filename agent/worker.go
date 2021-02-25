package agent

import (
	"context"
	"github.com/aos-dev/noah/proto"
	"google.golang.org/grpc"
)

type Worker struct {
}

func (w Worker) GetEndpoint(ctx context.Context, in *proto.GetEndpointRequest, opts ...grpc.CallOption) (*proto.GetEndpointReply, error) {
	panic("implement me")
}
