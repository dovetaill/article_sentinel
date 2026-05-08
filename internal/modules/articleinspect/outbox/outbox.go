package outbox

import "errors"

var ErrTaskOutboxDispatcherUnavailable = errors.New("task outbox dispatcher unavailable")
var ErrInvalidTaskInput = errors.New("invalid task input")

type TaskOutboxDispatchReport struct {
	Scanned      int
	Claimed      int
	Dispatched   int
	Retried      int
	DeadLettered int
	Failed       int
}
