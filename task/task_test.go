package task

import (
	"context"
	"github.com/aos-dev/noah/proto"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

const testGRPC = "localhost:7010"

func setupPortal(t *testing.T) *Portal {
	l, err := net.Listen("tcp", testGRPC)
	if err != nil {
		t.Error(err)
	}

	srv := grpc.NewServer()

	p, err := NewPortal()
	if err != nil {
		t.Error(err)
	}

	proto.RegisterNodeServer(srv, p)
	go func() {
		err = srv.Serve(l)
		if err != nil {
			t.Error(err)
		}
	}()

	return p
}

// This is not a really unit test, just for developing, SHOULD be removed.
func testWorker(t *testing.T) {
	p := setupPortal(t)

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		w, err := NewWorker(ctx, testGRPC)
		if err != nil {
			t.Error(err)
		}
		err = w.Connect(ctx)
		if err != nil {
			t.Error(err)
		}
	}

	copyFileJob := &proto.CopyDir{
		Src:     0,
		Dst:     1,
		SrcPath: "/tmp",
		DstPath: "/tmp",
	}
	content, err := protobuf.Marshal(copyFileJob)
	if err != nil {
		t.Error(err)
	}

	copyFileTask := &proto.Task{
		Id: uuid.NewString(),
		Endpoints: []*proto.Endpoint{
			{Type: "fs"},
			{Type: "fs"},
		},
		Job: &proto.Job{
			Id:      uuid.NewString(),
			Type:    TypeCopyDir,
			Content: content,
		},
	}
	err = p.Publish(ctx, copyFileTask)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(10 * time.Second)
}
