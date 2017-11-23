package timer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cocobao/howfw/netconn"
)

func timeout(time.Time, interface{}) {
	fmt.Println("123")
}

func TestTimer(t *testing.T) {
	ctx := context.Background()
	timerWheel := NewTimingWheel(ctx)
	timerWheel.AddTimer(netconn.IndId(), time.Now(), 5*time.Second, &OnTimeOut{
		Ctx:      ctx,
		Callback: timeout,
	})

	for {
	}
}
