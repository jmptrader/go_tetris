package utils

import (
	"fmt"
	"sync"
	"time"
)

type Frequecy struct {
	frequency int64
	data      map[interface{}]int64
	mu        sync.Mutex

	banDuration int64
	banList     map[interface{}]int64
	bmu         sync.Mutex
}

func NewFrequency(timesInMinute, banDurationInMinute int64) *Frequecy {
	f := &Frequecy{
		frequency:   timesInMinute,
		data:        make(map[interface{}]int64),
		banDuration: banDurationInMinute,
		banList:     make(map[interface{}]int64),
	}
	return f.init()
}

func (f *Frequecy) init() *Frequecy {
	go f.clear()
	go f.clearBan()
	return f
}

func (f *Frequecy) clear() {
	for {
		time.Sleep(time.Minute)
		func() {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.data = make(map[interface{}]int64)
		}()
	}
}

func (f *Frequecy) clearBan() {
	for {
		time.Sleep(5 * time.Second)
		func() {
			f.bmu.Lock()
			defer f.bmu.Unlock()
			tN := time.Now().Unix()
			for index, t := range f.banList {
				if tN-t > 60*f.banDuration {
					delete(f.banList, index)
				}
			}
		}()
	}
}

func (f *Frequecy) putIntoBan(index interface{}) {
	f.bmu.Lock()
	defer f.bmu.Unlock()
	f.banList[index] = time.Now().Unix()
}

func (f *Frequecy) Incr(index interface{}) error {
	// check if in ban list
	if err := func() error {
		f.bmu.Lock()
		defer f.bmu.Unlock()
		if f.banList[index] > 0 {
			return fmt.Errorf("正在黑名单中, 一会解禁")
		}
		return nil
	}(); err != nil {
		return err
	}

	// incr
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[index]++
	fmt.Println(f.data)
	if f.data[index] > f.frequency {
		f.putIntoBan(index)
		return fmt.Errorf("频率已超出每分钟%d次", f.frequency)
	}
	return nil
}
