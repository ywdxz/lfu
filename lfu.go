package lfu

import (
	"container/list"
	"sync"
)

var (
	placeholder = byte(0)
)

type LFU interface {
	Set(k string, v interface{})
	Get(k string) (v interface{}, ok bool)
	Evict(n int)
	Size() int
	// Print()
}

func New(cap int) LFU {

	if cap < 0 {
		cap = 0
	}

	return &cache{
		cap:      cap,
		kv:       make(map[string]*kvItem, cap),
		freqList: list.New(),
	}
}

type kvItem struct {
	k      string
	v      interface{}
	parent *list.Element
}

type freqNode struct {
	freq  int
	items map[*kvItem]interface{}
}

type cache struct {
	sync.Mutex
	cap      int
	kv       map[string]*kvItem
	freqList *list.List
}

func (c *cache) increment(item *kvItem) {

	//only update freqList
	if item == nil || item.parent == nil {
		return
	}

	cur := item.parent
	curNode := cur.Value.(*freqNode)

	if cur.Next() == nil || curNode.freq+1 != cur.Next().Value.(*freqNode).freq {

		item.parent = c.freqList.InsertAfter(&freqNode{
			freq:  curNode.freq + 1,
			items: map[*kvItem]interface{}{item: placeholder},
		}, cur)

	} else {
		nextNode := cur.Next().Value.(*freqNode)
		nextNode.items[item] = placeholder
	}

	delete(curNode.items, item)
	if len(curNode.items) == 0 {
		c.freqList.Remove(cur)
	}
	return
}

func (c *cache) Set(k string, v interface{}) {
	c.Mutex.Lock()
	c.set(k, v)
	c.Mutex.Unlock()
	return
}

func (c *cache) set(k string, v interface{}) {

	if c.cap <= 0 {
		return
	}

	if vv, ok := c.kv[k]; ok {
		//old
		vv.v = v
		c.increment(vv)
		return
	}

	if c.cap < len(c.kv)+1 {
		c.evict(len(c.kv) + 1 - c.cap)
	}

	var item *kvItem
	//new
	front := c.freqList.Front()
	if front == nil || front.Value.(*freqNode).freq != 1 {
		item = &kvItem{
			k: k,
			v: v,
		}
		node := &freqNode{
			freq:  1,
			items: map[*kvItem]interface{}{item: placeholder},
		}
		c.freqList.PushFront(node)
		item.parent = c.freqList.Front()
	} else {
		node := c.freqList.Front()
		item = &kvItem{
			k:      k,
			v:      v,
			parent: node,
		}
		node.Value.(*freqNode).items[item] = placeholder
	}
	c.kv[item.k] = item
	return
}

func (c *cache) Get(k string) (v interface{}, ok bool) {
	c.Mutex.Lock()
	v, ok = c.get(k)
	c.Mutex.Unlock()

	return
}

func (c *cache) get(k string) (v interface{}, ok bool) {

	item, ok := c.kv[k]
	if !ok {
		return
	}

	c.increment(item)
	v = item.v
	return
}

func (c *cache) Evict(n int) {
	c.Mutex.Lock()
	c.evict(n)
	c.Mutex.Unlock()
	return
}

func (c *cache) evict(n int) {

	for c.freqList.Len() > 0 && n > 0 {
		front := c.freqList.Front()
		frontNode := front.Value.(*freqNode)

		for item := range frontNode.items {
			if n <= 0 {
				break
			}
			delete(frontNode.items, item)
			delete(c.kv, item.k)
			n--
		}

		if len(frontNode.items) == 0 {
			c.freqList.Remove(front)
		}
	}
	return
}

func (c *cache) Size() (n int) {
	c.Mutex.Lock()
	n = c.size()
	c.Mutex.Unlock()
	return
}

func (c *cache) size() (n int) {
	n = len(c.kv)
	return
}

// func (c *cache) Print() {
// 	c.Mutex.Lock()
// 	defer c.Mutex.Unlock()

// 	for e := c.freqList.Front(); e != nil; e = e.Next() {
// 		node := e.Value.(*freqNode)
// 		fmt.Printf("[ %d - ", node.freq)

// 		for ite := range node.items {
// 			fmt.Printf("%+v,%+v ", ite.k, ite.v)
// 		}
// 		fmt.Printf("]\n")
// 	}
// 	return
// }
