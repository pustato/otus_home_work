package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	mx       sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheItem struct {
	key   Key
	value interface{}
}

func (cache *lruCache) Set(key Key, value interface{}) bool {
	cache.mx.Lock()
	defer cache.mx.Unlock()

	if listItem, ok := cache.items[key]; ok {
		ci := listItem.Value.(*cacheItem)
		ci.value = value
		cache.queue.MoveToFront(listItem)

		return true
	}

	ci := &cacheItem{key, value}
	listItem := cache.queue.PushFront(ci)
	cache.items[key] = listItem

	if cache.queue.Len() > cache.capacity {
		cache.purgeLastItem()
	}

	return false
}

func (cache *lruCache) Get(key Key) (interface{}, bool) {
	cache.mx.Lock()
	defer cache.mx.Unlock()

	listItem, ok := cache.items[key]
	if !ok {
		return nil, false
	}

	ci := listItem.Value.(*cacheItem)
	cache.queue.MoveToFront(listItem)

	return ci.value, true
}

func (cache *lruCache) Clear() {
	cache.mx.Lock()
	defer cache.mx.Unlock()

	cache.queue = NewList()
	cache.items = make(map[Key]*ListItem, cache.capacity)
}

func (cache *lruCache) purgeLastItem() {
	back := cache.queue.Back()
	ci := back.Value.(*cacheItem)
	cache.queue.Remove(back)

	delete(cache.items, ci.key)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
