package queueasynq

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	libasynq "github.com/hibiken/asynq"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

type enqueueRecorder struct {
	task *libasynq.Task
	opts []libasynq.Option
	info *libasynq.TaskInfo
	err  error
}

func (r *enqueueRecorder) Enqueue(task *libasynq.Task, opts ...libasynq.Option) (*libasynq.TaskInfo, error) {
	r.task = task
	r.opts = append([]libasynq.Option(nil), opts...)
	if r.info == nil {
		r.info = &libasynq.TaskInfo{ID: "task-1", Queue: "critical"}
	}
	return r.info, r.err
}

func TestNewTaskBuildsStableTaskNameAndPayload(t *testing.T) {
	payload := tasks.Payload{Source: "scheduler"}

	task, err := tasks.NewTask(payload)
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}
	if task.Type() != tasks.TypeRuntimeHeartbeat {
		t.Fatalf("task.Type() = %q, want %q", task.Type(), tasks.TypeRuntimeHeartbeat)
	}

	var got tasks.Payload
	if err := json.Unmarshal(task.Payload(), &got); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if got != payload {
		t.Fatalf("payload = %+v, want %+v", got, payload)
	}
}

func TestEnqueueHelperBuildsAsynqTask(t *testing.T) {
	recorder := &enqueueRecorder{}
	payload := tasks.Payload{Source: "scheduler"}

	info, err := EnqueueTask(recorder, "critical", payload)
	if err != nil {
		t.Fatalf("EnqueueTask() error = %v", err)
	}
	if info != recorder.info {
		t.Fatal("EnqueueTask() did not return underlying task info")
	}
	if recorder.task == nil {
		t.Fatal("EnqueueTask() did not enqueue a task")
	}
	if recorder.task.Type() != tasks.TypeRuntimeHeartbeat {
		t.Fatalf("enqueued task type = %q, want %q", recorder.task.Type(), tasks.TypeRuntimeHeartbeat)
	}

	queueName := ""
	for _, opt := range recorder.opts {
		if opt.Type() == libasynq.QueueOpt {
			queueName, _ = opt.Value().(string)
		}
	}
	if queueName != "critical" {
		t.Fatalf("queue option = %q, want %q", queueName, "critical")
	}
}

func TestRegisterHandlersReturnsMuxWithKnownTaskTypes(t *testing.T) {
	mux := RegisterHandlers(&bootstrap.Runtime{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	task, err := tasks.NewTask(tasks.Payload{Source: "worker"})
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}

	handler, pattern := mux.Handler(task)
	if handler == nil {
		t.Fatal("mux.Handler() returned nil handler")
	}
	if pattern != tasks.TypeRuntimeHeartbeat {
		t.Fatalf("pattern = %q, want %q", pattern, tasks.TypeRuntimeHeartbeat)
	}
	if err := mux.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("mux.ProcessTask() error = %v", err)
	}
}
