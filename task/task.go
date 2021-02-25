package task

import (
	"context"
	"sync"

	fs "github.com/aos-dev/go-service-fs/v2"
	"github.com/aos-dev/go-storage/v3/types"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"

	"github.com/aos-dev/noah/agent"
	"github.com/aos-dev/noah/proto"
)

const (
	TypeList uint32 = iota + 1
)

type Client struct {
	conn *nats.Conn
	sub  *nats.Subscription

	agent *agent.Worker

	storages sync.Map
}

func NewClient(addr string) (*Client, error) {
	nc, err := nats.Connect(addr)
	if err != nil {
		return nil, err
	}
	sub, err := nc.SubscribeSync("task")
	if err != nil {
		return nil, err
	}
	return &Client{conn: nc, sub: sub}, nil
}

func (c *Client) Publish(ctx context.Context, task *proto.Task) error {
	data, err := protobuf.Marshal(task)
	if err != nil {
		return err
	}

	return c.conn.Publish("task", data)
}

func (c *Client) Next(ctx context.Context) (*proto.Task, error) {
	var task *proto.Task

	msg, err := c.sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, err
	}

	err = protobuf.Unmarshal(msg.Data, task)
	if err != nil {
		panic("unmarshal failed")
	}

	return task, nil
}

func (c *Client) Handle(task *proto.Task) error {
	switch task.Type {
	case TypeList:
		var t *proto.List
		err := protobuf.Unmarshal(task.Content, t)
		if err != nil {
			panic("unmarshal failed")
		}
		return c.HandleList(t)
	}
	return nil
}

func (c *Client) GetStorage(id uint32) (types.Storager, error) {
	value, ok := c.storages.Load(id)
	if ok {
		return value.(types.Storager), nil
	}

	reply, err := c.agent.GetEndpoint(context.Background(), &proto.GetEndpointRequest{Id: id})
	if err != nil {
		return nil, err
	}

	var store types.Storager
	switch reply.Type {
	case fs.Type:
		store, err = fs.NewStorager()
	}
	if err != nil {
		return nil, err
	}
	c.storages.Store(id, store)

	return store, nil
}
