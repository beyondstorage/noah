package task

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aos-dev/go-toolbox/zapcontext"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	natsproto "github.com/nats-io/nats.go/encoders/protobuf"
	"go.uber.org/zap"

	"github.com/aos-dev/noah/proto"
)

type Portal struct {
	conn        *nats.EncodedConn
	nodes       []string
	nodeAddrMap map[string]string

	proto.UnimplementedNodeServer
}

func NewPortal() (p *Portal, err error) {
	srv, err := server.NewServer(&server.Options{
		Host:  "localhost",
		Port:  7100,
		Debug: true, // FIXME: allow used for developing
	})
	if err != nil {
		return
	}

	go func() {
		srv.ConfigureLogger()

		err = server.Run(srv)
		if err != nil {
			log.Printf("server run: %s", err)
		}
	}()

	if !srv.ReadyForConnections(time.Second) {
		panic(fmt.Errorf("server start too slow"))
	}

	p = &Portal{}
	p.nodeAddrMap = map[string]string{}

	conn, err := nats.Connect("localhost:7100")
	if err != nil {
		return
	}
	queue, err := nats.NewEncodedConn(conn, natsproto.PROTOBUF_ENCODER)
	if err != nil {
		return
	}
	p.conn = queue

	return p, nil
}

func (p *Portal) Register(ctx context.Context, request *proto.RegisterRequest) (*proto.RegisterReply, error) {
	log := zapcontext.From(ctx)

	log.Debug("got", zap.String("request", request.String()))
	p.nodes = append(p.nodes, request.Id)
	p.nodeAddrMap[request.Id] = request.Addr

	return &proto.RegisterReply{
		Addr:    "localhost:7100",
		Subject: fmt.Sprintf("node-%s", request.Id),
	}, nil
}

func (p *Portal) Upgrade(ctx context.Context, request *proto.UpgradeRequest) (*proto.UpgradeReply, error) {
	log := zapcontext.From(ctx)

	log.Debug("node addr map", zap.Reflect("map", p.nodeAddrMap))
	return &proto.UpgradeReply{
		NodeId:  p.nodes[0],
		Addr:    p.nodeAddrMap[p.nodes[0]],
		Subject: fmt.Sprintf("task-%s", request.TaskId),
	}, nil
}

func (p *Portal) mustEmbedUnimplementedAgentServer() {
	panic("implement me")
}

func (p *Portal) Publish(ctx context.Context, task *proto.Task) (err error) {
	log := zapcontext.From(ctx)

	for _, v := range p.nodes {
		log.Info("publish task to", zap.String("node", "node-"+v))
		err = p.conn.Publish(fmt.Sprintf("node-%s", v), task)
		if err != nil {
			return
		}
	}

	return
}
