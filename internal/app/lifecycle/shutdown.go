package lifecycle

import "errors"

// Closer 表示需要在退出阶段执行的关闭动作。
type Closer func() error

// Shutdown 按注册的逆序执行关闭动作，尽量回收全部资源。
func Shutdown(closers ...Closer) error {
	var err error
	for i := len(closers) - 1; i >= 0; i-- {
		if closers[i] == nil {
			continue
		}
		err = errors.Join(err, closers[i]())
	}
	return err
}
