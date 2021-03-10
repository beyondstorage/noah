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

func (rn *Runner) HandleCopyDir(ctx context.Context, msg protobuf.Message) error {
	_ = zapcontext.From(ctx)

	arg := msg.(*proto.CopyDir)

	store := rn.storages[arg.Src]

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

		err = rn.Async(ctx, &proto.Job{
			Id:      uuid.New().String(),
			Type:    TypeCopyFile,
			Content: content,
		})
		if err != nil {
			return err
		}

		err = rn.Await(ctx)
		if err != nil {
			return err
		}

	}
}

func (rn *Runner) HandleCopyFile(ctx context.Context, msg protobuf.Message) error {
	logger := zapcontext.From(ctx)

	arg := msg.(*proto.CopyFile)

	//src := rn.storages[arg.Src]
	//dst := rn.storages[arg.Dst]

	logger.Info("copy file",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}

func (rn *Runner) HandleCopySingleFile(ctx context.Context, msg protobuf.Message) error {
	logger := zapcontext.From(ctx)

	arg := msg.(*proto.CopySingleFile)

	logger.Info("copy single file",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}
func (rn *Runner) HandleCopyMultipartFile(ctx context.Context, msg protobuf.Message) error {
	logger := zapcontext.From(ctx)

	arg := msg.(*proto.CopyMultipartFile)

	// Send task and wait for response.
	logger.Info("copy multipart",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}

func (rn *Runner) HandleCopyMultipart(ctx context.Context, msg protobuf.Message) error {
	logger := zapcontext.From(ctx)

	arg := msg.(*proto.CopyMultipart)

	src := rn.storages[arg.Src]
	dst := rn.storages[arg.Dst]
	multipart, ok := dst.(types.Multiparter)
	if !ok {
		logger.Warn("storage does not implement Multiparter",
			zap.String("storage", dst.String()))
		return fmt.Errorf("not supported")
	}

	r, w := iowrap.Pipe()

	go func() {
		o := dst.Create(arg.DstPath, ps.WithMultipartID(arg.MultipartId))
		_, err := multipart.WriteMultipart(o, r, arg.Size, int(arg.Index))
		if err != nil {
			logger.Error("write multipart", zap.Error(err))
		}
	}()

	_, err := src.Read(arg.SrcPath, w)
	if err != nil {
		logger.Error("src read", zap.Error(err))
	}
	defer func() {
		err = r.Close()
		if err != nil {
			return
		}
	}()

	logger.Info("copy multipart",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}
