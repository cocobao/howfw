package timer

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func timeout(time.Time, interface{}) {
	fmt.Println("123")
}

func TestTimer(t *testing.T) {
	ctx := context.Background()
	timerWheel := NewTimingWheel(ctx)
	timerWheel.AddTimer(time.Now(), 5*time.Second, &OnTimeOut{
		Ctx:      ctx,
		Callback: timeout,
	})

	for {
	}
}
