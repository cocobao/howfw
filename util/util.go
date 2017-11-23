package util

import (
	"fmt"
	"runtime"
	"time"
)

func PrintStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	fmt.Printf("%s", buf[:n])
}

func Wait(sec time.Duration) {
	ticker := time.NewTicker(sec * time.Second)
	<-ticker.C
	ticker.Stop()
}
