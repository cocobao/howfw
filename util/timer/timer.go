package timer

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	tw   *TimingWheel
	one  sync.Once
	gctx context.Context
)

type OnTimeOut struct {
	Callback func(time.Time, interface{})
	Ctx      context.Context
}

type timerType struct {
	id         int64
	expiration time.Time
	interval   time.Duration
	timeout    *OnTimeOut
	index      int // for container/heap
}

func newTimer(id int64, when time.Time, interv time.Duration, to *OnTimeOut) *timerType {
	return &timerType{
		id:         id,
		expiration: when,
		interval:   interv,
		timeout:    to,
	}
}

func (t *timerType) isRepeat() bool {
	return int64(t.interval) > 0
}

type TimingWheel struct {
	timeOutChan chan *OnTimeOut
	timers      timerHeapType
	ticker      *time.Ticker
	wg          *sync.WaitGroup
	addChan     chan *timerType // add timer in loop
	cancelChan  chan int64      // cancel timer in loop
	sizeChan    chan int        // get size in loop
	ctx         context.Context
	cancel      context.CancelFunc
}

func GetTimingWheel() *TimingWheel {
	one.Do(func() {
		tw = &TimingWheel{
			timeOutChan: make(chan *OnTimeOut, 1024),
			timers:      make(timerHeapType, 0),
			ticker:      time.NewTicker(500 * time.Millisecond),
			wg:          &sync.WaitGroup{},
			addChan:     make(chan *timerType, 1024),
			cancelChan:  make(chan int64, 1024),
			sizeChan:    make(chan int),
		}
		gctx = context.Background()
		tw.ctx, tw.cancel = context.WithCancel(gctx)
		heap.Init(&tw.timers)
		tw.wg.Add(1)
		go func() {
			tw.start()
			tw.wg.Done()
		}()
	})

	return tw
}

//增加一个定时任务
func (tw *TimingWheel) AddTimer(when time.Time, interv time.Duration, to *OnTimeOut) int64 {
	if to == nil {
		return -1
	}
	id := time.Now().UnixNano()
	t := newTimer(id, when, interv, to)
	if tw == nil {
		fmt.Println("(E&(*#*(*(#$*((*U#*$(*&$#(")
	}
	tw.addChan <- t
	return id
}

func (tw *TimingWheel) Size() int {
	return <-tw.sizeChan
}

//删除一个定时任务
func (tw *TimingWheel) CancelTimer(timerID int64) {
	tw.cancelChan <- timerID
}

//暂停定时器
func (tw *TimingWheel) Stop() {
	tw.cancel()
	tw.wg.Wait()
}

func (tw *TimingWheel) getExpired() []*timerType {
	expired := make([]*timerType, 0)
	for tw.timers.Len() > 0 {
		timer := heap.Pop(&tw.timers).(*timerType)
		elapsed := time.Since(timer.expiration).Seconds()
		if elapsed > 1.0 {
			// log.Warnf("elapsed %f\n", elapsed)
		}
		if elapsed > 0.0 {
			//出栈已经超时的
			expired = append(expired, timer)
			continue
		} else {
			//由于定时器是降序排序的，有一个还没有到超时时间，后面的都是没有超时的
			heap.Push(&tw.timers, timer)
			break
		}
	}
	return expired
}

func (tw *TimingWheel) update(timers []*timerType) {
	if timers != nil {
		for _, t := range timers {
			//是否需要重复定时
			if t.isRepeat() {
				//重新压栈
				heap.Push(&tw.timers, t)
			}
		}
	}
}

func (tw *TimingWheel) start() {
	for {
		select {
		//删除一个定时任务
		case timerID := <-tw.cancelChan:
			index := tw.timers.getIndexByID(timerID)
			if index >= 0 {
				heap.Remove(&tw.timers, index)
			}

		case tw.sizeChan <- tw.timers.Len():

			//暂停运行
		case <-tw.ctx.Done():
			tw.ticker.Stop()
			return

			//增加一个定时任务
		case timer := <-tw.addChan:
			heap.Push(&tw.timers, timer)

			//检测是否有定时器到超时时间
		case <-tw.ticker.C:
			timers := tw.getExpired()
			now := time.Now()

			for _, t := range timers {
				expiration := t.expiration.Add(t.interval)

				if now.Unix() >= expiration.Unix() {
					t.timeout.Callback(now, nil)
					t.expiration = now
				}
			}
			tw.update(timers)
		}
	}
}
