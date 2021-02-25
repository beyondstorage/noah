package task

import (
	"context"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"

	"github.com/aos-dev/noah/proto"
)

type Client struct {
	conn *nats.Conn

	sub *nats.Subscription
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
