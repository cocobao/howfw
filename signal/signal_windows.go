package signal

import "fmt"
import "os"
import "os/signal"
import "syscall"
import "github.com/cocobao/log"

type signalHandler func(s os.Signal, arg interface{})

type signalSet struct {
	m map[os.Signal]signalHandler
}

func signalSetNew() *signalSet {
	ss := new(signalSet)
	ss.m = make(map[os.Signal]signalHandler)
	return ss
}

func (set *signalSet) register(s os.Signal, handler signalHandler) {
	if _, found := set.m[s]; !found {
		set.m[s] = handler
	}
}

func (set *signalSet) handle(sig os.Signal, arg interface{}) (err error) {
	if _, found := set.m[sig]; found {
		set.m[sig](sig, arg)
		return nil
	} else {
		return fmt.Errorf("No handler available for signal %v", sig)
	}
}

func GracefullyStopSever(release func()) {
	ss := signalSetNew()
	//ctrl + c
	interruptHandler := func(s os.Signal, arg interface{}) {
		if st := shutdown(); st {
			log.Info("windows server shutdown success!")
			release()
			os.Exit(0)
		} else {
			log.Info("server shutdown failed!")
		}
	}
	ss.register(syscall.SIGINT, interruptHandler)
	for {
		c := make(chan os.Signal)
		signal.Notify(c)
		sig := <-c
		err := ss.handle(sig, nil)
		if err != nil {
			log.Infof("windows unknown signal received: %v, %v\n", sig, err)
		}
	}
}
