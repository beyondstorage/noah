package task

import (
	"context"
	"github.com/aos-dev/go-storage/v3/types"
	"github.com/aos-dev/noah/proto"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"log"
)

func (a *Agent) HandleCopyDir(ctx context.Context, msg protobuf.Message) error {
	arg := msg.(*proto.CopyDir)

	store := a.storages[arg.Src]

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

		err = a.Publish(ctx, &proto.Job{
			Id:      uuid.New().String(),
			Type:    TypeCopyFile,
			Content: content,
		})
		if err != nil {
			return err
		}
	}
}

func (a *Agent) HandleCopyFile(ctx context.Context, msg protobuf.Message) error {
	arg := msg.(*proto.CopyFile)

	//src := a.storages[arg.Src]
	//dst := a.storages[arg.Dst]

	log.Printf("copy file from %s to %s", arg.SrcPath, arg.DstPath)
	return nil
}

func (a *Agent) HandleCopySingleFile(ctx context.Context, msg protobuf.Message) error {
	arg := msg.(*proto.CopySingleFile)

	log.Printf("copy single file from %s to %s", arg.SrcPath, arg.DstPath)
	return nil
}
func (a *Agent) HandleCopyMultipartFile(ctx context.Context, msg protobuf.Message) error {
	arg := msg.(*proto.CopyMultipartFile)

	log.Printf("copy multipart from %s to %s", arg.SrcPath, arg.DstPath)
	return nil
}

func (a *Agent) HandleCopyMultipart(ctx context.Context, msg protobuf.Message) error {
	arg := msg.(*proto.CopyMultipart)

	log.Printf("copy multipart from %s to %s", arg.SrcPath, arg.DstPath)
	return nil
}
