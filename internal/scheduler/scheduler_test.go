package scheduler

import (
	"context"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/robfig/cron/v3"
)

type enqueueRecorder struct {
	calls    int
	payloads []tasks.Payload
	err      error
}

type outboxRelayRecorder struct {
	calls      int
	limits     []int
	dispatched int
	err        error
}

func (r *enqueueRecorder) EnqueueRuntimeHeartbeat(payload tasks.Payload) error {
	r.calls++
	r.payloads = append(r.payloads, payload)
	return r.err
}

func (r *outboxRelayRecorder) RelayPendingArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error) {
	_ = ctx
	r.calls++
	r.limits = append(r.limits, limit)
	return r.dispatched, r.err
}

func TestRegisterJobsAddsCronEntries(t *testing.T) {
	c := cron.New()
	rt := &bootstrap.Runtime{
		Config: &config.Config{
			Scheduler: config.SchedulerConfig{
				Enabled: true,
				Spec:    "@every 1m",
			},
			Queue: config.QueueConfig{
				Outbox: config.OutboxConfig{
					Enabled:   true,
					RelaySpec: "@every 15s",
					BatchSize: 20,
				},
			},
		},
	}

	if err := RegisterJobs(c, rt, &enqueueRecorder{}, &outboxRelayRecorder{}); err != nil {
		t.Fatalf("RegisterJobs() error = %v", err)
	}

	if got := len(c.Entries()); got != 2 {
		t.Fatalf("len(c.Entries()) = %d, want %d", got, 2)
	}
}

func TestScheduledJobOnlyEnqueuesTask(t *testing.T) {
	recorder := &enqueueRecorder{}

	job := NewRuntimeHeartbeatJob(nil, recorder)
	job()

	if recorder.calls != 1 {
		t.Fatalf("calls = %d, want %d", recorder.calls, 1)
	}
	if len(recorder.payloads) != 1 {
		t.Fatalf("len(payloads) = %d, want %d", len(recorder.payloads), 1)
	}
	if recorder.payloads[0].Source != "scheduler" {
		t.Fatalf("payload.Source = %q, want %q", recorder.payloads[0].Source, "scheduler")
	}
}

func TestArticleInspectTaskOutboxRelayJobDispatchesPendingMessages(t *testing.T) {
	recorder := &outboxRelayRecorder{dispatched: 3}

	job := NewArticleInspectTaskOutboxRelayJob(nil, recorder, 20)
	job()

	if recorder.calls != 1 {
		t.Fatalf("calls = %d, want %d", recorder.calls, 1)
	}
	if len(recorder.limits) != 1 || recorder.limits[0] != 20 {
		t.Fatalf("limits = %#v, want %#v", recorder.limits, []int{20})
	}
}
