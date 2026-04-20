package logger

import (
	"time"
)

type rotationTicker interface {
	C() <-chan time.Time
	Stop()
}

type stdTicker struct {
	ticker *time.Ticker
}

func (t *stdTicker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *stdTicker) Stop() {
	t.ticker.Stop()
}

type fileRotator interface {
	Rotate() error
}

var timeNow = time.Now

func newStdTicker(d time.Duration) rotationTicker {
	return &stdTicker{ticker: time.NewTicker(d)}
}

// startDailyRotation 在日期变化后主动触发一次日志轮转。
func startDailyRotation(now func() time.Time, newTicker func(time.Duration) rotationTicker, rotator fileRotator) func() error {
	ticker := newTicker(time.Minute)
	done := make(chan struct{})
	lastRotateDay := dayKey(now())

	go func() {
		for {
			select {
			case <-done:
				return
			case currentTime := <-ticker.C():
				currentDay := dayKey(currentTime)
				if currentDay == lastRotateDay {
					continue
				}

				_ = rotator.Rotate()
				lastRotateDay = currentDay
			}
		}
	}()

	return func() error {
		close(done)
		ticker.Stop()
		return nil
	}
}

func dayKey(t time.Time) string {
	return t.Format("2006-01-02")
}
