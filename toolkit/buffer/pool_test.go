package buffer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolNew(t *testing.T) {
	bufferCap := uint32(10)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when creating a new pool: %s (bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}
	if pool == nil {
		t.Fatalf("Could not create a new pool")
	}
	if pool.BufferCap() != bufferCap {
		t.Fatalf("Inconsistent buffer cap, expected: %d, actual: %d", bufferCap, pool.BufferCap())
	}
	if pool.MaxBufferNumber() != maxBufferNumber {
		t.Fatalf("Inconsistent maxBufferNumber, expected: %d, actual: %d", maxBufferNumber, pool.MaxBufferNumber())
	}
	if pool.BufferNumber() != 1 {
		t.Fatalf("Inconsistent buffer number, expected: %d, actual: %d", 1, pool.BufferNumber())
	}
	_, err = NewPool(0, 1)
	if err == nil {
		t.Fatalf("No error when creating a buffer pool with zero buffer cap")
	}
	_, err = NewPool(1, 0)
	if err == nil {
		t.Fatalf("No error when creating a buffer pool with zero maxBufferNumber")
	}
}

func addExtraDatum(pool Pool, datum interface{}) chan error {
	sign := make(chan error, 1)
	go func() {
		sign <- pool.Put(datum)
	}()
	return sign
}

func TestPoolPut(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when creating a buffer pool: %s (bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}
	dataLen := bufferCap * maxBufferNumber
	data := make([]uint32, dataLen)
	for i := uint32(0); i < dataLen; i++ {
		data[i] = i
	}
	var count, datum uint32
	for _, datum := range data {
		err := pool.Put(datum)
		if err != nil {
			t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, datum)
		}
		count++
		if pool.Total() != uint64(count) {
			t.Fatalf("Inconsistent data total, expected: %d, actual: %d", count, pool.Total())
		}
		expectedBufferNumber := count / uint32(bufferCap)
		if count%uint32(bufferCap) != 0 {
			expectedBufferNumber++
		}
		if pool.BufferNumber() != expectedBufferNumber {
			t.Fatalf("Inconsistent buffer number, expected: %d, actual: %d (count: %d)",
				expectedBufferNumber, pool.BufferNumber(), count)
		}
	}
	select {
	case err := <-addExtraDatum(pool, datum):
		if err != nil {
			t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, datum)
		} else {
			t.Fatal("It can still put a datum to the full buffer pool")
		}
	case <-time.After(time.Millisecond):
		t.Logf("Timeout! could not put data to the full buffer pool")
	}
	pool.Close()
	err = pool.Put(datum)
	if err == nil {
		t.Fatalf("It can still put data to the closed buffer pool (datum: %d)", datum)
	}
}

func TestPoolPutInParallel(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when creating a buffer pool: %s (bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}
	dataLen := bufferCap * maxBufferNumber
	data := make([]uint32, dataLen)
	for i := uint32(0); i < dataLen; i++ {
		data[i] = i
	}
	var count uint32
	testingFunc := func(datum interface{}) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			err := pool.Put(datum)
			if err != nil {
				t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, datum)
			}
			atomic.AddUint32(&count, 1)
			currentCount := atomic.LoadUint32(&count)
			if uint64(currentCount) > pool.Total() {
				t.Fatalf("Inconsistent data total, old > new %d > %d", currentCount, pool.Total())
			}
		}
	}
	t.Run("Put in parallel 1", func(t *testing.T) {
		for _, datum := range data[:dataLen/2] {
			t.Run(fmt.Sprintf("Datum=%d", datum), testingFunc(datum))
		}
	})
	t.Run("Put in parallel 2", func(t *testing.T) {
		for _, datum := range data[dataLen/2:] {
			t.Run(fmt.Sprintf("Datum=%d", datum), testingFunc(datum))
		}
	})
	datum := dataLen
	select {
	case err := <-addExtraDatum(pool, datum):
		if err != nil {
			t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, datum)
		} else {
			t.Fatal("It can still put a datum to the full buffer pool")
		}
	case <-time.After(time.Millisecond):
		t.Logf("Timeout could not put data to the full buffer pool")
	}
	pool.Close()
}

func getExtraDatum(pool Pool) chan error {
	sign := make(chan error, 1)
	go func() {
		_, err := pool.Get()
		sign <- err
	}()
	return sign
}

func TestPoolGet(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when creating a buffer pool: %s (bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}
	dataLen := uint32(bufferCap * maxBufferNumber)
	for i := uint32(0); i < dataLen; i++ {
		pool.Put(i)
	}
	count := dataLen
	expectedBufferNumber := maxBufferNumber
	var datum uint32
	var ok bool
	for i := uint32(0); i < dataLen; i++ {
		d, err := pool.Get()
		if err != nil {
			t.Fatalf("An error occurs when getting a datum for the buffer pool: %s", err)
		}
		datum, ok = d.(uint32)
		if !ok {
			t.Fatalf("Inconsistent datum type, expected: %T, actual: %T", datum, d)
		}
		if datum >= dataLen {
			t.Fatalf("datum out of range, expected: [0, %d], actual: %d", dataLen, datum)
		}
		count--
		if pool.Total() != uint64(count) {
			t.Fatalf("Inconsistent data total, expected: %d, actual: %d", count, pool.Total())
		}
		if pool.BufferNumber() != expectedBufferNumber {
			t.Fatalf("Inconsistent buffer number, expected: %d, actual: %d (count: %d)", expectedBufferNumber, pool.BufferNumber(), count)
		}
	}
	select {
	case err := <-getExtraDatum(pool):
		if err != nil {
			t.Fatalf("An error occurs when getting a datum for the buffer pool: %s", err)
		} else {
			t.Fatal("It can still get a datum from the empty buffer pool")
		}
	case <-time.After(time.Millisecond):
		t.Logf("Timeout! could not get data from the empty buffer pool")
	}
	datum = 0
	pool.Put(datum)
	pool.Close()
	_, err = pool.Get()
	if err == nil {
		t.Fatal("It can still get datum from the closed buffer pool")
	}
}

func TestPoolGetInParallel(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when creating a buffer pool: %s (bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}
	dataLen := uint32(bufferCap * maxBufferNumber)
	for i := uint32(0); i < dataLen; i++ {
		pool.Put(i)
	}
	count := dataLen
	testingFunc := func(t *testing.T) {
		t.Parallel()
		d, err := pool.Get()
		if err != nil {
			t.Fatalf("An error occurs when getting a datum from the buffer pool: %s", err)
		}
		datum, ok := d.(uint32)
		if !ok {
			t.Fatalf("Inconsistent datum type, expected: %T, actual: %T", datum, d)
		}
		if datum >= dataLen {
			t.Fatalf("datum out of range: [0, %d], actual: %d", dataLen, datum)
		}
		atomic.AddUint32(&count, ^uint32(0))
		currentCount := atomic.LoadUint32(&count)
		if uint64(currentCount) < pool.Total() {
			t.Fatalf("Inconsistent data total: %d < %d old < new", currentCount, pool.Total())
		}
	}
	t.Run("Get in parallel 1", func(t *testing.T) {
		min := uint32(0)
		max := dataLen / 2
		for i := min; i < max; i++ {
			t.Run(fmt.Sprintf("Index=%d", i), testingFunc)
		}
	})
	t.Run("Get in parallel 2", func(t *testing.T) {
		min := dataLen / 2
		max := dataLen
		for i := min; i < max; i++ {
			t.Run(fmt.Sprintf("Index=%d", i), testingFunc)
		}
	})
	select {
	case err := <-getExtraDatum(pool):
		if err != nil {
			t.Fatalf("An error occurs when getting a datum from the buffer pool: %s", err)
		} else {
			t.Fatal("It can still get a datum from the empty buffer pool")
		}
	case <-time.After(time.Millisecond):
		t.Logf("Timeout! could not get data from the empty buffer pool")
	}
	pool.Close()
}

func TestPoolPutAndgetInParllel(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when creating a buffer pool: %s (bufferCap: %d, maxNumber: %d)", err, bufferCap, maxBufferNumber)
	}
	dataLen := uint32(bufferCap * maxBufferNumber)
	maxPuttingNumber := dataLen + uint32(rand.Int63n(20))
	maxGettingNumber := dataLen + uint32(rand.Int63n(20))
	puttingCount := maxPuttingNumber
	gettingCount := maxGettingNumber
	marks := make([]uint32, maxPuttingNumber)
	var lock sync.Mutex
	t.Run("All in Parallel", func(t *testing.T) {
		t.Run("Put 1", func(t *testing.T) {
			t.Parallel()
			begin := uint32(0)
			end := maxPuttingNumber / 2
			for i := begin; i < end; i++ {
				if pool.Total() == uint64(dataLen) {
					datum := dataLen
					select {
					case err := <-addExtraDatum(pool, datum):
						if err != nil {
							t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, datum)
						} else {
							t.Fatal("It can still put a datum to the full buffer pool")
						}
					case <-time.After(time.Millisecond):
						t.Logf("Timeout! could not put data to the full buffer pool")
					}
					continue
				}
				err := pool.Put(i)
				if err != nil {
					t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, i)
				}
				atomic.AddUint32(&puttingCount, ^uint32(0))
			}
		})
		t.Run("Put 2", func(t *testing.T) {
			t.Parallel()
			begin := maxPuttingNumber / 2
			end := maxPuttingNumber
			for i := begin; i < end; i++ {
				if pool.Total() == uint64(dataLen) {
					datum := dataLen
					select {
					case err := <-addExtraDatum(pool, datum):
						if err != nil {
							t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, datum)
						} else {
							t.Fatal("It can still put a datum to the full buffer pool")
						}
					case <-time.After(time.Millisecond):
						t.Logf("Timeout! could not put data to the full buffer pool")
					}
					continue
				}
				err := pool.Put(i)
				if err != nil {
					t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, i)
				}
				atomic.AddUint32(&puttingCount, ^uint32(0))
			}
		})
		t.Run("Get 1", func(t *testing.T) {
			t.Parallel()
			max := dataLen/2 + 1
			for i := uint32(0); i < max; i++ {
				if pool.Total() == 0 {
					select {
					case err := <-getExtraDatum(pool):
						if err != nil {
							t.Fatalf("An error occurs when getting a datum from the empty buffer pool: %s", err)
						}
					case <-time.After(time.Microsecond):
						t.Logf("Timeout! could get data from the empty buffer pool")
					}
					continue
				}
				d, err := pool.Get()
				if err != nil {
					t.Fatalf("An error occurs when getting a datum from the buffer pool: %s", err)
				}
				if d == nil && atomic.LoadUint32(&puttingCount) == 0 && pool.Total() != 0 {
					t.Fatalf("Get an empty datum (total: %d)", pool.Total())
				}
				atomic.AddUint32(&gettingCount, ^uint32(0))
				if d != nil {
					datum := d.(uint32)
					lock.Lock()
					marks[int(datum)]++
					lock.Unlock()
				}
			}
		})
		t.Run("Get 2", func(t *testing.T) {
			t.Parallel()
			max := dataLen/2 + 2
			for i := uint32(0); i < max; i++ {
				if pool.Total() == 0 {
					select {
					case err := <-getExtraDatum(pool):
						if err != nil {
							t.Fatalf("An error occurs when getting a datum from the empty buffer pool: %s", err)
						}
					case <-time.After(time.Microsecond):
						t.Logf("Timeout! could get data from the empty buffer pool")
					}
					continue
				}
				d, err := pool.Get()
				if err != nil {
					t.Fatalf("An error occurs when getting a datum from the buffer pool: %s", err)
				}
				if d == nil && atomic.LoadUint32(&puttingCount) == 0 && pool.Total() != 0 {
					t.Fatalf("Get an empty datum (total: %d)", pool.Total())
				}
				atomic.AddUint32(&gettingCount, ^uint32(0))
				if d != nil {
					datum := d.(uint32)
					lock.Lock()
					marks[int(datum)]++
					lock.Unlock()
				}
			}
		})
	})
	for i, m := range marks {
		if m > 1 {
			t.Fatalf("Got the number more than once: %d", i)
		}
	}
	pool.Close()
}

func TestPoolCloseInParallel(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when creating a new buffer pool: %s (bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}
	dataLen := uint32(bufferCap * maxBufferNumber)
	maxNumber := dataLen / 2
	t.Run("Put", func(t *testing.T) {
		t.Parallel()
		for i := uint32(0); i < maxNumber; i++ {
			err := pool.Put(i)
			if err != nil && !pool.Closed() {
				t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)", err, i)
			}
		}
	})
	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		for i := uint32(0); i < maxNumber; i++ {
			_, err := pool.Get()
			if err != nil && !pool.Closed() {
				t.Fatalf("An error occurs when getting a datum from the buffer pool: %s (index: %d)", err, i)
			}
		}
	})
	t.Run("Close", func(t *testing.T) {
		t.Parallel()
		time.Sleep(time.Millisecond)
		ok := pool.Close()
		if !ok {
			t.Fatal("Could not close the buffer pool")
		}
		if !pool.Closed() {
			t.Fatalf("Inconsistent buffer pool status, expected: %v, actual: %v", true, pool.Closed())
		}
		ok = pool.Close()
		if ok {
			t.Fatal("It can still close the closed buffer pool")
		}
		if !pool.Closed() {
			t.Fatalf("Inconsistent buffer pool status, expected: %v, actual: %v", true, pool.Closed())
		}
	})
}
