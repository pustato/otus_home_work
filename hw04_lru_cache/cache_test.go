package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)

		c.Set("item1", 1)
		c.Set("item2", 2)
		c.Set("item3", 3)
		c.Set("item4", 4)
		c.Set("item5", 5)

		state := [...]struct {
			key Key
			ok  bool
		}{
			{"item1", false},
			{"item2", false},
			{"item3", true},
			{"item4", true},
			{"item5", true},
		}

		for _, si := range state {
			_, ok := c.Get(si.key)
			require.Equal(t, si.ok, ok)
		}
	})

	t.Run("purge oldest", func(t *testing.T) {
		c := NewCache(5)

		c.Set("item1", 1)
		c.Set("item2", 2)
		c.Set("item3", 3)
		c.Set("item4", 4)
		c.Set("item5", 5)

		for _, key := range [...]Key{"item3", "item2", "item1", "item3", "item2", "item1"} {
			_, ok := c.Get(key)
			require.True(t, ok)
		}

		c.Set("item6", 6)

		_, ok := c.Get("item4")
		require.False(t, ok)

		for val, key := range [...]Key{"item3", "item2", "item1", "item5", "item6", "item3"} {
			wasInCache := c.Set(key, val)
			require.True(t, wasInCache)
		}

		c.Set("item7", 7)

		_, ok = c.Get("item2")
		require.False(t, ok)
	})

	t.Run("test clear", func(t *testing.T) {
		c := NewCache(3)

		c.Set("item1", 1)
		c.Set("item2", 2)
		c.Set("item3", 3)

		c.Clear()

		for _, key := range [...]Key{"item1", "item2", "item3"} {
			_, ok := c.Get(key)
			require.False(t, ok)
		}
	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
