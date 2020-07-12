package lru

import (
	"container/list"
)

//字典(map)，存储键和值的映射关系。这样根据某个键(key)查找对应的值(value)的复杂是O(1)，在字典中插入一条记录的复杂度也是O(1)
//双向链表(double linked list)实现的队列。将所有的值放到双向链表中，这样，当访问到某个值时，将其移动到队尾的复杂度是O(1)，在队尾新增一条记录以及删除一条记录的复杂度均为O(1)

type Cache struct {
	maxBytes  int64      //允许使用的最大内存
	nbytes    int64      //当前已使用的内存
	ll        *list.List //双向链表
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) //某条记录被移除时的回调函数
}

//双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}

	return
}

func (c *Cache) Add(key string, value Value) {
	//如果键存在，则更新对应节点的值，并将该节点移到队头。
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		//不存在则是新增场景，首先队尾添加新节点 &entry{key, value}, 并字典中添加 key 和节点的映射关系。
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//更新 c.nbytes，如果超过了设定的最大值 c.maxBytes
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

//缓存淘汰
func (c *Cache) RemoveOldest() {
	//取队首节点，从链表删除
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		//从字典中c.cache删除该节点的映射关系
		delete(c.cache, kv.key)
		//更新当前所用的内存c.nbytes
		c.nbytes -= int64(len(kv.key)) + int64((kv.value.Len()))
		//如果回调函数onEvicted不为nil,则调回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
