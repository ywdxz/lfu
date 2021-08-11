package lfu

type LFU interface {
	Set(k string, v interface{})
	Get(k string) (v interface{}, ok bool)
}

type listNode struct {
	key        string
	value      interface{}
	prev, next *listNode
	count      int
}

type twoWayList struct {
	head, tail *listNode
	len        int
}

func (t *twoWayList) addToHead(cur *listNode) {
	switch {
	case t.len == 0:
		t.head = cur
		t.tail = cur
		t.len = 1
	default:
		t.head.prev = cur
		cur.next = t.head
		t.head = cur
		t.len++
	}
}

func (t *twoWayList) delNode(cur *listNode) {
	switch {
	case t.len == 0:
	case t.len == 1:
		t.head = nil
		t.tail = nil
		t.len = 0
		cur.next = nil
		cur.prev = nil
	case cur == t.head:
		t.head = t.head.next
		t.head.prev = nil
		t.len--
		cur.next = nil
		cur.prev = nil
	case cur == t.tail:
		t.tail = t.tail.prev
		t.tail.next = nil
		t.len--
		cur.next = nil
		cur.prev = nil
	default:
		cur.next.prev = cur.prev
		cur.prev.next = cur.next
		t.len--
		cur.next = nil
		cur.prev = nil
	}
}

func (t *twoWayList) delTailNode() (node *listNode) {
	node = t.tail
	switch {
	case t.len == 0:
	case t.len == 1:
		t.head = nil
		t.tail = nil
		t.len = 0
		node.prev = nil
		node.next = nil
	default:
		t.tail = t.tail.prev
		t.tail.next = nil
		t.len--
		node.prev = nil
		node.next = nil
	}
	return
}

func (t *twoWayList) ifEmpty() bool {
	if t.len <= 0 {
		return true
	}
	return false
}

type lfuCache struct {
	cap, len int
	hash1    map[string]*listNode
	hash2    map[int]*twoWayList
	minCount int
}

func New(cap int) LFU {

	if cap < 0 {
		cap = 0
	}

	return &lfuCache{
		cap:      cap,
		len:      0,
		minCount: 0,
		hash1:    make(map[string]*listNode, cap),
		hash2:    make(map[int]*twoWayList, cap),
	}
}

func (l *lfuCache) updateNode(cur *listNode) {

	li, _ := l.hash2[cur.count]

	li.delNode(cur)

	if li.ifEmpty() {
		delete(l.hash2, cur.count)
		if l.minCount == cur.count {
			l.minCount++
		}
	}

	cur.count++

	li, ok := l.hash2[cur.count]
	if !ok {
		li = &twoWayList{}
		l.hash2[cur.count] = li
	}
	li.addToHead(cur)
}

// eliminateNode 未更新 lfuCache.count
func (l *lfuCache) eliminateNode() {
	li, _ := l.hash2[l.minCount]
	v := li.delTailNode()
	delete(l.hash1, v.key)
	if li.ifEmpty() {
		delete(l.hash2, v.count)
	}
	l.len--
}

func (l *lfuCache) addNewNode(cur *listNode) {
	cur.count = 1
	l.hash1[cur.key] = cur
	li, ok := l.hash2[cur.count]
	if !ok {
		li = &twoWayList{}
		l.hash2[cur.count] = li
	}
	li.addToHead(cur)
	l.minCount = cur.count
	l.len++
}

func (l *lfuCache) Set(k string, v interface{}) {

	if l.cap <= 0 {
		return
	}

	if v, ok := l.hash1[k]; ok {
		v.value = v
		l.updateNode(v)
		return
	}

	if l.len == l.cap {
		l.eliminateNode()
	}

	l.addNewNode(&listNode{
		key:   k,
		value: v,
	})
}

func (l *lfuCache) Get(k string) (v interface{}, ok bool) {
	if v, ok := l.hash1[k]; ok {
		l.updateNode(v)
		return v.value, true
	}
	return nil, false
}
