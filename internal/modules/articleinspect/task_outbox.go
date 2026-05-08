package articleinspect

import outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"

var ErrTaskOutboxDispatcherUnavailable = outboxpkg.ErrTaskOutboxDispatcherUnavailable

type TaskOutboxDispatchReport = outboxpkg.TaskOutboxDispatchReport
