package util

import "sync/atomic"

type AtomicBool int32

func NewAtomicBool(val bool) *AtomicBool {
	b := new(AtomicBool)
	b.Set(val)
	return b
}

func (b *AtomicBool) Set(val bool) {
	if val {
		atomic.StoreInt32((*int32)(b), 1)
	} else {
		atomic.StoreInt32((*int32)(b), 0)
	}
}

func (b *AtomicBool) Value() bool {
	return atomic.LoadInt32((*int32)(b))&1 == 1
}

func (b *AtomicBool) Toggle() bool {
	return atomic.AddInt32((*int32)(b), 1)&1 == 0
}
