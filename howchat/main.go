package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cocobao/howfw/util/timer"

	"github.com/cocobao/howfw/howchat/handle"
	"github.com/cocobao/log"
)

var (
	timerWheel *timer.TimingWheel
)

func console() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">>>")
		data, _, _ := reader.ReadLine()
		handle.CommandHandle(strings.Split(string(data), " "))
	}
}

func main() {
	log.NewLogger("", log.LoggerLevelDebug)

	console()
}
