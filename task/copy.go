package task

import (
	"context"
	"fmt"

	ps "github.com/aos-dev/go-storage/v3/pairs"
	"github.com/aos-dev/go-storage/v3/pkg/iowrap"
	"github.com/aos-dev/go-storage/v3/types"
	"github.com/aos-dev/go-toolbox/zapcontext"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aos-dev/noah/proto"
)

func (a *Agent) HandleCopyDir(ctx context.Context, msg protobuf.Message) error {
	_ = zapcontext.From(ctx)

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
	log := zapcontext.From(ctx)

	arg := msg.(*proto.CopyFile)

	//src := a.storages[arg.Src]
	//dst := a.storages[arg.Dst]

	log.Info("copy file",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}

func (a *Agent) HandleCopySingleFile(ctx context.Context, msg protobuf.Message) error {
	log := zapcontext.From(ctx)

	arg := msg.(*proto.CopySingleFile)

	log.Info("copy single file",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}
func (a *Agent) HandleCopyMultipartFile(ctx context.Context, msg protobuf.Message) error {
	log := zapcontext.From(ctx)

	arg := msg.(*proto.CopyMultipartFile)

	// Send task and wait for response.
	log.Info("copy multipart",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}

func (a *Agent) HandleCopyMultipart(ctx context.Context, msg protobuf.Message) error {
	log := zapcontext.From(ctx)

	arg := msg.(*proto.CopyMultipart)

	src := a.storages[arg.Src]
	dst := a.storages[arg.Dst]
	mulipart, ok := dst.(types.Multiparter)
	if !ok {
		log.Warn("storage does not implement Multiparter",
			zap.String("storage", dst.String()))
		return fmt.Errorf("not supported")
	}

	r, w := iowrap.Pipe()

	go func() {
		o := dst.Create(arg.DstPath, ps.WithMultipartID(arg.MultipartId))
		_, err := mulipart.WriteMultipart(o, r, arg.Size, int(arg.Index))
		if err != nil {
			log.Error("write multipart", zap.Error(err))
		}
	}()

	_, err := src.Read(arg.SrcPath, w)
	if err != nil {
		log.Error("src read", zap.Error(err))
	}
	defer func() {
		err = r.Close()
		if err != nil {
			return
		}
	}()

	log.Info("copy multipart",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}
