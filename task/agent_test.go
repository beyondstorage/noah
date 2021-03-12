package task

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aos-dev/go-storage/v3/pkg/credential"
	"github.com/aos-dev/go-toolbox/zapcontext"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/google/uuid"

	"github.com/aos-dev/noah/proto"
)

func setupPortal(t *testing.T) *Portal {
	p, err := NewPortal(context.Background(), PortalConfig{
		Host:      "localhost",
		GrpcPort:  7000,
		QueuePort: 7010,
	})
	if err != nil {
		t.Error(err)
	}

	return p
}

// This is not a really unit test, just for developing, SHOULD be removed.
func testWorker(t *testing.T) {
	p := setupPortal(t)

	ctx := context.Background()
	logger := zapcontext.From(ctx)

	for i := 0; i < 3; i++ {
		w, err := NewWorker(ctx, WorkerConfig{
			Host:       "localhost",
			PortalAddr: "localhost:7000",
		})
		if err != nil {
			t.Error(err)
		}
		err = w.Connect(ctx)
		if err != nil {
			t.Error(err)
		}
	}

	copyFileJob := &proto.CopyDir{
		Src:       0,
		Dst:       1,
		SrcPath:   "",
		DstPath:   "",
		Recursive: true,
	}
	content, err := protobuf.Marshal(copyFileJob)
	if err != nil {
		t.Error(err)
	}

	ak, sk := os.Getenv("QS_ACCESS_KEY"), os.Getenv("QS_SECRET_KEY")
	cred := credential.NewHmac(ak, sk)
	copyFileTask := &proto.Task{
		Id: uuid.NewString(),
		Endpoints: []*proto.Endpoint{
			{Type: "fs", Pairs: []*proto.Pair{{Key: "work_dir", Value: "/tmp/"}}},
			{Type: "qingstor", Pairs: []*proto.Pair{
				{Key: "work_dir", Value: "/tmp3/d/"},
				{Key: "name", Value: "lance-qsctl2"},
				{Key: "location", Value: "gd2"},
				{Key: "credential", Value: cred.String()},
			}},
		},
		Job: &proto.Job{
			Id:      uuid.NewString(),
			Type:    TypeCopyDir,
			Content: content,
		},
	}

	logger.Info("before first publish")
	err = p.Publish(ctx, copyFileTask)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(30 * time.Second)
}
