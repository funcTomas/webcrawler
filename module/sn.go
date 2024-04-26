package module

import (
	"math"
	"sync"
)

type SNGenerator interface {
	Start() uint64
	Max() uint64
	Next() uint64
	CycleCount() uint64
	Get() uint64
}

func NewSNGenerator(start uint64, max uint64) SNGenerator {
	if max == 0 {
		max = math.MaxUint64
	}
	return &mySnGenerator{
		start: start,
		max:   max,
		next:  start,
	}
}

type mySnGenerator struct {
	start      uint64
	max        uint64
	next       uint64
	cycleCount uint64
	lock       sync.RWMutex
}

func (gen *mySnGenerator) Start() uint64 {
	return gen.start
}

func (gen *mySnGenerator) Max() uint64 {
	return gen.max
}

func (gen *mySnGenerator) Next() uint64 {
	gen.lock.RLock()
	defer gen.lock.RUnlock()
	return gen.next
}

func (gen *mySnGenerator) CycleCount() uint64 {
	gen.lock.RLock()
	defer gen.lock.RUnlock()
	return gen.cycleCount
}

func (gen *mySnGenerator) Get() uint64 {
	gen.lock.Lock()
	defer gen.lock.Unlock()
	id := gen.next
	if id == gen.max {
		gen.next = gen.start
		gen.cycleCount++
	} else {
		gen.next++
	}
	return id
}
