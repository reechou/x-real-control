package utils

import (
	"sync"
	"time"
)

type TimingWheel struct {
	sync.Mutex

	interval time.Duration
	ticker   *time.Ticker
	stop     chan struct{}

	cs      []chan struct{}
	buckets int

	pos int
}

func NewTimingWheel(interval time.Duration, buckets int) *TimingWheel {
	w := &TimingWheel{
		interval: interval,
		ticker:   time.NewTicker(interval),
		stop:     make(chan struct{}),
		buckets:  buckets,
	}
	w.cs = make([]chan struct{}, buckets)
	for i := range w.cs {
		w.cs[i] = make(chan struct{})
	}

	go w.run()

	return w
}

func (self *TimingWheel) Stop() {
	close(self.stop)
}

func (self *TimingWheel) Check(id int64) <-chan struct{} {
	self.Lock()
	idx := int(id) % self.buckets
	b := self.cs[idx]
	self.Unlock()

	return b
}

func (self *TimingWheel) run() {
	for {
		select {
		case <-self.ticker.C:
			self.onTicker()
		case <-self.stop:
			self.ticker.Stop()
			return
		}
	}
}

func (self *TimingWheel) onTicker() {
	self.Lock()
	lastC := self.cs[self.pos]
	self.cs[self.pos] = make(chan struct{})
	self.pos = (self.pos + 1) % self.buckets
	self.Unlock()

	close(lastC)
}
