package articleinspect

import "errors"

var ErrTaskOutboxDispatcherUnavailable = errors.New("task outbox dispatcher unavailable")

type TaskOutboxDispatchReport struct {
	Scanned      int
	Claimed      int
	Dispatched   int
	Retried      int
	DeadLettered int
	Failed       int
}
