package tests

import (
	"sync/atomic"
	"testing"
)

func TestAtomicBool(t *testing.T) {
	flag := atomic.Bool{}

	flag.Store(false)

	swap := flag.CompareAndSwap(true, true)

	t.Log(swap, flag.Load())
}
