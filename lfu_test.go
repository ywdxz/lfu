package lfu

import (
	"container/list"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLFU(t *testing.T) {
	cache := New(2)

	cache.Set("a", 1)
	assert.Equal(t, 1, cache.Size())
	cache.Set("b", 2)
	assert.Equal(t, 2, cache.Size())
	v, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	cache.Evict(1)
	assert.Equal(t, 1, cache.Size())
	v, ok = cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	v, ok = cache.Get("b")
	assert.False(t, ok)
	assert.Nil(t, v)
	cache.Set("c", 3)
	assert.Equal(t, 2, cache.Size())
	v, ok = cache.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, v)
	cache.Set("d", 4)
	assert.Equal(t, 2, cache.Size())
	v, ok = cache.Get("c")
	assert.False(t, ok)
	assert.Nil(t, v)
	v, ok = cache.Get("d")
	assert.True(t, ok)
	assert.Equal(t, 4, v)
	cache.Evict(10)
	assert.Equal(t, 0, cache.Size())
}

func TestCache_Set(t *testing.T) {
	cache := &cache{
		cap:      2,
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
	}

	// set "a"
	cache.Set("a", 1)
	assert.Equal(t, 1, cache.kv["a"].v)
	assert.Equal(t, 1, cache.freqList.Len())
	frontNode := cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 1, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok := frontNode.items[cache.kv["a"]]
	assert.True(t, ok)

	// set "a" again
	cache.Set("a", 2)
	assert.Equal(t, 2, cache.kv["a"].v)
	assert.Equal(t, 1, cache.freqList.Len())
	frontNode = cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 2, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok = frontNode.items[cache.kv["a"]]
	assert.True(t, ok)

	// set "b"
	cache.Set("b", 1)
	assert.Equal(t, 1, cache.kv["b"].v)
	assert.Equal(t, 2, cache.freqList.Len())
	frontNode = cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 1, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok = frontNode.items[cache.kv["b"]]
	assert.True(t, ok)
	nextNode := cache.freqList.Front().Next().Value.(*freqNode)
	assert.Equal(t, 2, nextNode.freq)
	assert.Equal(t, 1, len(nextNode.items))
	_, ok = nextNode.items[cache.kv["a"]]
	assert.True(t, ok)

	// set "c" should evict "b"
	cache.Set("c", 1)
	assert.Equal(t, 2, len(cache.kv))
	assert.Equal(t, 2, cache.freqList.Len())
	frontNode = cache.freqList.Front().Value.(*freqNode)
	assert.Equal(t, 1, frontNode.freq)
	assert.Equal(t, 1, len(frontNode.items))
	_, ok = frontNode.items[cache.kv["c"]]
	assert.True(t, ok)
}

func TestCache_Get(t *testing.T) {
	cache := &cache{
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
		cap:      0,
	}

	v, ok := cache.Get("c")
	assert.False(t, ok)
	assert.Nil(t, v)
	assert.Equal(t, 0, cache.freqList.Len())
	assert.Equal(t, 0, len(cache.kv))

	cache.Set("a", 1)
	cache.Set("b", 2)

	v, ok = cache.Get("a")

	// assert.True(t, ok)
	// assert.Equal(t, 1, v)
	assert.False(t, ok)
	assert.Nil(t, v)

}

func TestCache_Size(t *testing.T) {
	cache := &cache{
		kv:       make(map[string]*kvItem),
		freqList: list.New(),
	}

	assert.Equal(t, 0, cache.Size())

	cache.Set("a", 1)
	assert.Equal(t, 0, cache.Size())

	cache.Set("b", 1)
	assert.Equal(t, 0, cache.Size())

	cache.Evict(10)
	assert.Equal(t, 0, cache.Size())
}
