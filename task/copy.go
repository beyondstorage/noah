package task

import (
	"context"
	"fmt"

	ps "github.com/aos-dev/go-storage/v3/pairs"
	"github.com/aos-dev/go-storage/v3/pkg/iowrap"
	"github.com/aos-dev/go-storage/v3/types"
	"github.com/aos-dev/go-toolbox/zapcontext"
	protobuf "github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/aos-dev/noah/proto"
)

const defaultMultipartThreshold = 128 * 1024 * 1024

func (rn *Runner) HandleCopyDir(ctx context.Context, msg protobuf.Message) error {
	logger := rn.logger
	arg := msg.(*proto.CopyDir)

	store := rn.storages[arg.Src]

	it, err := store.List(arg.SrcPath)
	if err != nil {
		logger.Error("storage list failed",
			zap.Error(err),
			zap.String("store", store.String()),
			zap.String("path", arg.SrcPath))
		return err
	}

	for {
		o, err := it.Next()
		if err == types.IterateDone {
			break
		}
		if err != nil {
			logger.Error("get next object failed",
				zap.Error(err),
				zap.String("store", store.String()))
			return err
		}

		// if obj is dir and not recursive, skip directly
		if o.GetMode().IsDir() && !arg.Recursive {
			continue
		}

		job := proto.NewJob()
		// set job attr separately for dir and file
		if o.GetMode().IsDir() {
			content, err := protobuf.Marshal(&proto.CopyDir{
				Src:       arg.Src,
				Dst:       arg.Dst,
				SrcPath:   o.Path,
				DstPath:   o.Path,
				Recursive: true,
			})
			if err != nil {
				panic("marshal failed")
			}
			job.Type = TypeCopyDir
			job.Content = content
		} else {
			content, err := protobuf.Marshal(&proto.CopyFile{
				Src:     arg.Src,
				Dst:     arg.Dst,
				SrcPath: o.Path,
				DstPath: o.Path,
			})
			if err != nil {
				panic("marshal failed")
			}
			job.Type = TypeCopyFile
			job.Content = content
		}

		err = rn.Async(ctx, &job)
		if err != nil {
			logger.Error("async job failed",
				zap.Error(err),
				zap.String("job", job.String()),
				zap.String("store", store.String()))
			return err
		}
	}

	if err = rn.Await(ctx); err != nil {
		logger.Error("await job failed",
			zap.Error(err),
			zap.String("runner job", rn.j.String()),
			zap.String("store", store.String()))
		return err
	}
	return nil
}

func (rn *Runner) HandleCopyFile(ctx context.Context, msg protobuf.Message) error {
	logger := rn.logger
	arg := msg.(*proto.CopyFile)

	store := rn.storages[arg.Src]

	obj, err := store.Stat(arg.SrcPath)
	if err != nil {
		return err
	}
	size, ok := obj.GetContentLength()
	if !ok {
		return fmt.Errorf("object %s size not set", arg.SrcPath)
	}

	job := proto.NewJob()
	if _, ok := store.(types.Multiparter); ok && size > defaultMultipartThreshold {
		content, err := protobuf.Marshal(&proto.CopyMultipartFile{
			Src:     arg.Src,
			Dst:     arg.Dst,
			SrcPath: arg.SrcPath,
			DstPath: arg.DstPath,
			Size:    size,
		})
		if err != nil {
			panic("marshal failed")
		}

		job.Type = TypeCopyMultipartFile
		job.Content = content
	} else {
		content, err := protobuf.Marshal(&proto.CopySingleFile{
			Src:     arg.Src,
			Dst:     arg.Dst,
			SrcPath: arg.SrcPath,
			DstPath: arg.DstPath,
			Size:    size,
		})
		if err != nil {
			panic("marshal failed")
		}

		job.Type = TypeCopySingleFile
		job.Content = content
	}

	if err := rn.Sync(ctx, &job); err != nil {
		logger.Error("await job failed",
			zap.Error(err),
			zap.String("runner job", rn.j.String()),
			zap.String("store", store.String()))
		return err
	}

	logger.Info("copy file",
		zap.String("from", arg.SrcPath),
		zap.String("to", arg.DstPath))
	return nil
}

func (rn *Runner) HandleCopySingleFile(ctx context.Context, msg protobuf.Message) error {
	logger := rn.logger

	arg := msg.(*proto.CopySingleFile)

	src := rn.storages[arg.Src]
	dst := rn.storages[arg.Dst]

	r, w := iowrap.Pipe()

	go func() {
		_, err := dst.Write(arg.DstPath, r, arg.Size)
		if err != nil {
			logger.Error("write multipart: %v", zap.Error(err))
		}
	}()

	_, err := src.Read(arg.SrcPath, w)
	if err != nil {
		logger.Error("src read: %v", zap.Error(err))
	}
	defer func() {
		err = r.Close()
		if err != nil {
			return
		}
	}()

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
	logger := rn.logger

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
		return err
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
