package count

import (
	"sync"
)

type Map struct {
	sync.Map
}

var list *Map

func New(from, to int) *Map {
	list = &Map{}
	var zero int
	for from <= to {
		list.Store(from, zero)
		from++
	}
	return list
}

func (l *Map) MergeFromSource(source *sync.Map) {
	source.Range(func(key, value any) bool {
		currentCount, _ := l.Load(key.(int))
		currentCount = currentCount.(int) + 1
		return true
	})
}
