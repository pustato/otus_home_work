package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("test push", func(t *testing.T) {
		l := NewList()

		middle := l.PushBack("middle") // ["middle"]
		require.Equal(t, "middle", middle.Value)
		require.Equal(t, "middle", l.Back().Value)
		require.Equal(t, "middle", l.Front().Value)
		require.Equal(t, 1, l.Len())

		front := l.PushFront("front") // ["front", "middle"]
		require.Equal(t, "front", front.Value)
		require.Equal(t, "front", l.Front().Value)
		require.Equal(t, "middle", l.Back().Value)
		require.Equal(t, 2, l.Len())

		back := l.PushBack("back") // ["front", "middle", "back"]
		require.Equal(t, "back", back.Value)
		require.Equal(t, "front", l.Front().Value)
		require.Equal(t, "back", l.Back().Value)
		require.Equal(t, 3, l.Len())
	})

	t.Run("test remove", func(t *testing.T) {
		l := NewList()

		front := l.PushBack("front")   // ["front"]
		middle := l.PushBack("middle") // ["front", "middle"]
		back := l.PushBack("back")     // ["front", "middle", "back"]

		l.Remove(front) // ["middle", "back"]
		require.Nil(t, middle.Prev)
		require.Equal(t, "middle", l.Front().Value)
		require.Equal(t, "back", l.Back().Value)
		require.Equal(t, 2, l.Len())

		l.Remove(back) // ["middle"]
		require.Nil(t, middle.Prev)
		require.Nil(t, middle.Next)
		require.Equal(t, "middle", l.Front().Value)
		require.Equal(t, "middle", l.Back().Value)
		require.Equal(t, 1, l.Len())

		l.Remove(middle) // []
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
		require.Equal(t, 0, l.Len())
	})

	t.Run("test move", func(t *testing.T) {
		l := NewList()

		l.PushBack("front")            // ["front"]
		middle := l.PushBack("middle") // ["front", "middle"]
		l.PushBack("another middle")   // ["front", "middle", "another middle"]
		back := l.PushBack("back")     // ["front", "middle", "another middle", "back"]

		l.MoveToFront(middle) // ["middle", "front", "another middle", "back"]
		require.Equal(t, "middle", l.Front().Value)
		require.Equal(t, "back", l.Back().Value)

		l.MoveToFront(back) // ["back", "middle", "front", "another middle"]
		require.Equal(t, "back", l.Front().Value)
		require.Equal(t, "another middle", l.Back().Value)
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}
