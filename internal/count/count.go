package count

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"sync"
	"syscall"
	"time"
)

type Info interface {
	Done()
	EncodeBin() []byte
	GetMax() int
	GetMin() int
	MergeFromSource(*sync.Map)
	Pause()
	Print()
	Reset()
	Resume()
	setCount()
	Start(int)
	LoadMap(key int) (value any, ok bool)
}

type info struct {
	sync.Map
	signal chan syscall.Signal
	wg     *sync.WaitGroup
	max    int
	min    int
}

var list *info

var pause bool

func (l *info) GetMax() int {
	return l.max
}

func (l *info) GetMin() int {
	return l.min
}

func (l *info) LoadMap(key int) (value any, ok bool) {
	return l.Load(key)
}

func New(from, to int) Info {
	list = &info{}
	list.min = from
	list.max = to

	var zero int
	for from <= to {
		list.Store(from, zero)
		from++
	}
	return list
}

func (l *info) MergeFromSource(source *sync.Map) {
	source.Range(func(key, value any) bool {
		currentCount, _ := l.Load(key.(int))
		currentCount = currentCount.(int) + 1
		return true
	})
}

func (l *info) Print() {
	l.Range(walk)
}

func (l *info) Reset() {
	var zero int
	from, to := l.min, l.max
	for from <= to {
		l.Delete(from)
		l.Store(from, zero)
		from++
	}
}

func walk(key, value any) bool {
	fmt.Println("Key =", key, "count =", value.(int))
	return true
}

func (l *info) Start(thNum int) {
	l.wg = &sync.WaitGroup{}
	l.wg.Add(thNum)
	l.signal = make(chan syscall.Signal)

	for i := 1; i <= thNum; i++ {
		go l.setCount()
	}
}

func (l *info) Pause() {
	pause = true
}

func (l *info) Resume() {
	pause = false
}

func (l *info) setCount() {
	defer l.wg.Done()
	rand.Seed(time.Now().UnixNano())
	for {
		select {
		case <-l.signal:
			return
		default:
			if pause {
				continue
			}
			randNum := rand.Intn(l.max-l.min+1) + l.min
			currentCount, ok := l.Load(randNum)
			if !ok {
				l.Store(randNum, 1)
			} else {
				currentCount = currentCount.(int) + 1
				l.Store(randNum, currentCount)
			}
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
	}
}

func (l *info) Done() {
	l.signal <- syscall.SIGHUP
}

func (l *info) EncodeBin() []byte {
	tempMap := make(map[int]int)
	l.Range(func(key, value any) bool {
		tempMap[key.(int)] = value.(int)
		return true
	})
	buf := bytes.NewBuffer(nil)
	_ = gob.NewEncoder(buf).Encode(&tempMap)
	return buf.Bytes()
}

func Decode(data []byte) (*info, error) {
	i := &info{}
	decodeMap := make(map[int]int)
	buf := bytes.NewBuffer(data)
	err := gob.NewDecoder(buf).Decode(&decodeMap)
	for key, value := range decodeMap {
		i.Store(key, value)
	}
	return i, err
}
