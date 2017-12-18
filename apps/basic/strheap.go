package main

import (
	"container/heap"
	"fmt"
)

// An IntHeap is a min-heap of ints.
type StrHeap []string

func (h StrHeap) Len() int           { return len(h) }
func (h StrHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h StrHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *StrHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(string))
}

func (h *StrHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// This example inserts several ints into an IntHeap, checks the minimum,
// and removes them in order of priority.
func doStrHeap() {
	h := &StrHeap{"好", "23", "Golang"}
	heap.Init(h)
	heap.Push(h, "0")
	fmt.Printf("first element: %s\n", (*h)[0])
	for h.Len() > 0 {
		fmt.Printf("%s ", heap.Pop(h))
	}
	// Output:
	// minimum: 0
	// 0 23 Golang 好
}
