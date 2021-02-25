package task

import (
	"context"
	"github.com/aos-dev/go-storage/v3/types"
	"github.com/aos-dev/noah/proto"
	protobuf "github.com/golang/protobuf/proto"
)

func (c *Client) HandleCopyDir(ctx context.Context, arg *proto.CopyDir) error {
	store, err := c.GetStorage(arg.Src)
	if err != nil {
		return err
	}

	it, err := store.List(arg.SrcPath)
	if err != nil {
		return err
	}

	for {
		o, err := it.Next()
		if err == types.IterateDone {
			return nil
		}
		if err != nil {
			return err
		}

		content, err := protobuf.Marshal(&proto.CopyFile{
			Src:     arg.Src,
			Dst:     arg.Dst,
			SrcPath: o.Path,
			DstPath: o.Path,
		})
		if err != nil {
			panic("marshal failed")
		}

		err = c.Publish(ctx, &proto.Task{
			Id:      0,
			Type:    0,
			Content: content,
		})
		if err != nil {
			return err
		}
	}
}
