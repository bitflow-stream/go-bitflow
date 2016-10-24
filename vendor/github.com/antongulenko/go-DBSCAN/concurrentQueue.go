package dbscan

import (
	"sync/atomic"
	"unsafe"
)

type queueNode struct {
	Value uint
	Next  unsafe.Pointer
}

type ConcurrentQueue_InsertOnly struct {
	Head unsafe.Pointer
	Size uint64
}

func NewConcurrentQueue_InsertOnly() *ConcurrentQueue_InsertOnly {
	var q = new(ConcurrentQueue_InsertOnly)
	q.Head = unsafe.Pointer(new(queueNode))
	return q
}

func (self *ConcurrentQueue_InsertOnly) Add(value uint) {
	var node = new(queueNode)
	node.Value = value
	node.Next = self.Head

	for atomic.CompareAndSwapPointer(&self.Head, node.Next, unsafe.Pointer(node)) == false {
		node.Next = self.Head
	}
	atomic.AddUint64(&self.Size, 1)
}

func (self *ConcurrentQueue_InsertOnly) Slice() []uint {
	var (
		result = make([]uint, 0, self.Size)
		node   = (*queueNode)(self.Head)
	)
	// for node.Next != nil {
	for i := uint64(0); i < self.Size; i += 1 {
		result = append(result, node.Value)
		node = (*queueNode)(node.Next)
	}
	return result
}
