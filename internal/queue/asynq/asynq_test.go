package queueasynq

import (
	"context"
	"encoding/json"
	"errors"
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

func TestArticleInspectQueueBuildsPayload(t *testing.T) {
	recorder := &enqueueRecorder{}
	payload := tasks.ArticleInspectTaskPayload{
		TaskID:        88,
		OrgID:         100,
		TriggerSource: "manual",
		OperatorID:    9,
		OperatorName:  "alice",
	}

	info, err := EnqueueArticleInspectTask(recorder, "critical", payload)
	if err != nil {
		t.Fatalf("EnqueueArticleInspectTask() error = %v", err)
	}
	if info != recorder.info {
		t.Fatal("EnqueueArticleInspectTask() did not return underlying task info")
	}
	if recorder.task == nil {
		t.Fatal("EnqueueArticleInspectTask() did not enqueue a task")
	}
	if recorder.task.Type() != tasks.TypeArticleInspectRunTask {
		t.Fatalf("task.Type() = %q, want %q", recorder.task.Type(), tasks.TypeArticleInspectRunTask)
	}
	var (
		queueName string
		taskID    string
	)
	for _, opt := range recorder.opts {
		switch opt.Type() {
		case libasynq.QueueOpt:
			queueName, _ = opt.Value().(string)
		case libasynq.TaskIDOpt:
			taskID, _ = opt.Value().(string)
		}
	}
	if queueName != "critical" {
		t.Fatalf("queue option = %q, want %q", queueName, "critical")
	}
	if taskID != "articleinspect-task-88" {
		t.Fatalf("task id option = %q, want %q", taskID, "articleinspect-task-88")
	}

	decoded, err := tasks.DecodeArticleInspectTaskPayload(recorder.task)
	if err != nil {
		t.Fatalf("DecodeArticleInspectTaskPayload() error = %v", err)
	}
	if decoded != payload {
		t.Fatalf("payload = %+v, want %+v", decoded, payload)
	}
}

func TestArticleInspectDispatcherIgnoresTaskIDConflict(t *testing.T) {
	recorder := &enqueueRecorder{err: libasynq.ErrTaskIDConflict}
	dispatcher := NewArticleInspectTaskDispatcher(recorder, "critical")
	if dispatcher == nil {
		t.Fatal("NewArticleInspectTaskDispatcher() = nil, want dispatcher")
	}

	err := dispatcher.DispatchArticleInspectTask(context.Background(), tasks.ArticleInspectTaskPayload{
		TaskID:        88,
		OrgID:         100,
		TriggerSource: "manual",
	})
	if err != nil {
		t.Fatalf("DispatchArticleInspectTask() error = %v, want nil", err)
	}
}

func TestArticleInspectDispatcherPropagatesEnqueueError(t *testing.T) {
	wantErr := errors.New("redis down")
	recorder := &enqueueRecorder{err: wantErr}
	dispatcher := NewArticleInspectTaskDispatcher(recorder, "critical")
	if dispatcher == nil {
		t.Fatal("NewArticleInspectTaskDispatcher() = nil, want dispatcher")
	}

	err := dispatcher.DispatchArticleInspectTask(context.Background(), tasks.ArticleInspectTaskPayload{
		TaskID:        88,
		OrgID:         100,
		TriggerSource: "manual",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("DispatchArticleInspectTask() error = %v, want %v", err, wantErr)
	}
}

func TestArticleInspectQueueHandlerDispatchesExecutor(t *testing.T) {
	orig := newArticleInspectExecutorFn
	t.Cleanup(func() {
		newArticleInspectExecutorFn = orig
	})

	recorder := &articleInspectExecutorRecorder{}
	newArticleInspectExecutorFn = func(rt *bootstrap.Runtime) articleInspectExecutor {
		return recorder
	}

	mux := RegisterHandlers(&bootstrap.Runtime{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	task, err := tasks.NewArticleInspectTask(tasks.ArticleInspectTaskPayload{
		TaskID:        88,
		OrgID:         100,
		TriggerSource: "manual",
	})
	if err != nil {
		t.Fatalf("NewArticleInspectTask() error = %v", err)
	}

	if err := mux.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("mux.ProcessTask() error = %v", err)
	}
	if recorder.calls != 1 {
		t.Fatalf("executor calls = %d, want %d", recorder.calls, 1)
	}
	if recorder.lastPayload.TaskID != 88 || recorder.lastPayload.OrgID != 100 {
		t.Fatalf("executor payload = %+v, want task/org 88/100", recorder.lastPayload)
	}
}

type articleInspectExecutorRecorder struct {
	calls       int
	lastPayload tasks.ArticleInspectTaskPayload
}

func (r *articleInspectExecutorRecorder) ExecuteTask(ctx context.Context, payload tasks.ArticleInspectTaskPayload) error {
	r.calls++
	r.lastPayload = payload
	return nil
}
