package vt100

import (
	"sync"
)

// BoundedList 是一个有容量上限的列表，当达到上限时自动丢弃最早的元素
type BoundedList[T any] struct {
	mu       sync.RWMutex
	items    []T
	capacity int
	start    int // 循环数组的起始索引
	size     int // 当前元素数量
}

// New 创建一个新的有界列表
func NewBoundedList[T any](capacity int) *BoundedList[T] {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}
	return &BoundedList[T]{
		items:    make([]T, capacity),
		capacity: capacity,
		start:    0,
		size:     0,
	}
}

// Push 添加一个元素到列表末尾，如果达到容量上限则丢弃最早的元素
func (b *BoundedList[T]) Push(item T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.size < b.capacity {
		// 还有空间，直接添加
		idx := (b.start + b.size) % b.capacity
		b.items[idx] = item
		b.size++
	} else {
		// 已满，覆盖最早的元素（start 位置）
		b.items[b.start] = item
		b.start = (b.start + 1) % b.capacity
	}
}

// Pop 移除并返回最早的元素
func (b *BoundedList[T]) Pop() (T, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.size == 0 {
		var zero T
		return zero, false
	}

	item := b.items[b.start]
	var zero T
	b.items[b.start] = zero // 帮助 GC 回收
	b.start = (b.start + 1) % b.capacity
	b.size--

	return item, true
}

// Get 获取指定索引的元素（0 表示最早的元素）
func (b *BoundedList[T]) Get(index int) (T, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if index < 0 || index >= b.size {
		var zero T
		return zero, false
	}

	idx := (b.start + index) % b.capacity
	return b.items[idx], true
}

// First 获取最早的元素
func (b *BoundedList[T]) First() (T, bool) {
	return b.Get(0)
}

// Last 获取最新的元素
func (b *BoundedList[T]) Last() (T, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.size == 0 {
		var zero T
		return zero, false
	}

	idx := (b.start + b.size - 1) % b.capacity
	return b.items[idx], true
}

// Len 返回当前元素数量
func (b *BoundedList[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Cap 返回列表容量
func (b *BoundedList[T]) Cap() int {
	return b.capacity
}

// IsFull 检查列表是否已满
func (b *BoundedList[T]) IsFull() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size == b.capacity
}

// IsEmpty 检查列表是否为空
func (b *BoundedList[T]) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size == 0
}

// Clear 清空列表
func (b *BoundedList[T]) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 清空所有元素帮助 GC
	for i := 0; i < b.capacity; i++ {
		var zero T
		b.items[i] = zero
	}
	b.start = 0
	b.size = 0
}

// Values 返回所有元素（按从旧到新的顺序）
func (b *BoundedList[T]) Values() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]T, b.size)
	for i := 0; i < b.size; i++ {
		idx := (b.start + i) % b.capacity
		result[i] = b.items[idx]
	}
	return result
}

// ForEach 遍历所有元素（从旧到新）
func (b *BoundedList[T]) ForEach(fn func(index int, value T)) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := 0; i < b.size; i++ {
		idx := (b.start + i) % b.capacity
		fn(i, b.items[idx])
	}
}

func (b *BoundedList[T]) Copy() *BoundedList[T] {
	n := NewBoundedList[T](b.capacity)
	b.ForEach(func(index int, value T) {
		n.Push(value)
	})
	return n
}
