package rhmonitor4go

import "sync/atomic"

type StaticsMainter struct {
	ordInsertCount uint64
	ordActionCount uint64
	offsetOrdCount uint64
}

func (stat *StaticsMainter) OrderInsertInc() uint64 {
	return atomic.AddUint64(&stat.ordInsertCount, 1)
}

func (stat *StaticsMainter) OrderActionInc() uint64 {
	return atomic.AddUint64(&stat.ordActionCount, 1)
}

func (stat *StaticsMainter) OffsetOrderInc() uint64 {
	return atomic.AddUint64(&stat.offsetOrdCount, 1)
}
