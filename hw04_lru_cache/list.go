package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value      interface{}
	Next, Prev *ListItem
}

type list struct {
	size        int
	front, back *ListItem
}

func (l *list) Len() int {
	return l.size
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: nil, Prev: nil}
	l.putFront(item)

	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: nil, Prev: nil}
	l.putBack(item)

	return item
}

func (l *list) Remove(i *ListItem) {
	switch {
	case i.Next == nil && i.Prev == nil:
		l.back, l.front, l.size = nil, nil, 0

	case i.Next == nil:
		newBack := i.Prev
		l.back, newBack.Next, l.size = newBack, nil, l.size-1

	case i.Prev == nil:
		newFront := i.Next
		l.front, newFront.Prev, l.size = newFront, nil, l.size-1

	default:
		prev, next := i.Prev, i.Next
		prev.Next, next.Prev, l.size = next, prev, l.size-1
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if l.front == i {
		return
	}

	l.Remove(i)
	l.putFront(i)
}

func (l *list) putFront(item *ListItem) {
	item.Next = l.front

	if l.front != nil {
		l.front.Prev = item
	}

	if l.back == nil {
		l.back = item
	}

	l.front = item
	l.size++
}

func (l *list) putBack(item *ListItem) {
	item.Prev = l.back

	if l.back != nil {
		l.back.Next = item
	}

	if l.front == nil {
		l.front = item
	}

	l.back = item
	l.size++
}

func NewList() List {
	return new(list)
}
