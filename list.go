package sqlconvert

// ListwItem represents an item in the linked list.
type ListwItem struct {
	value interface{}
	prev  *ListwItem
	next  *ListwItem
}

// NewListwItem creates a new ListwItem with default values.
func NewListwItem() *ListwItem {
	return &ListwItem{
		value: nil,
		prev:  nil,
		next:  nil,
	}
}

// ListW represents a wrapped list structure.
type ListW struct {
	first *ListwItem
	last  *ListwItem
	count int
}

// NewListW creates a new ListW instance.
func NewListW() *ListW {
	return &ListW{
		first: nil,
		last:  nil,
		count: 0,
	}
}

// DeleteAll removes all elements from the list.
func (l *ListW) DeleteAll() {
	current := l.first
	for current != nil {
		next := current.next
		current = next
	}
	l.first = nil
	l.last = nil
	l.count = 0
}

// DeleteLast removes the last item from the list.
func (l *ListW) DeleteLast() {
	l.DeleteSince(l.GetLast())
}

// DeleteSince removes elements starting from the specified item, including it.
func (l *ListW) DeleteSince(since *ListwItem) {
	if since == nil {
		return
	}

	current := since
	if since == l.first {
		l.first = nil
	}

	l.last = since.prev
	if l.last != nil {
		l.last.next = nil
	}

	for current != nil {
		next := current.next
		current = next
		l.count--
	}
}

// Add adds a new item to the list.
func (l *ListW) Add(value interface{}) {
	item := NewListwItem()
	item.value = value

	if l.first == nil {
		l.first = item
	} else {
		l.last.next = item
		item.prev = l.last
	}

	l.last = item
	l.count++
}

// GetFirst returns the first item in the list.
func (l *ListW) GetFirst() *ListwItem {
	return l.first
}

// GetLast returns the last item in the list.
func (l *ListW) GetLast() *ListwItem {
	return l.last
}

// GetCount returns the total number of items in the list.
func (l *ListW) GetCount() int {
	return l.count
}

// ListwmItem represents an item in the linked list.
type ListwmItem struct {
	Value1 interface{}
	Value2 interface{}
	Value3 interface{}
	Value4 interface{}
	Value5 interface{}
	Value6 interface{}
	IValue int
	Prev   *ListwmItem
	Next   *ListwmItem
}

// NewListwmItem creates a new ListwmItem with default values.
func NewListwmItem() *ListwmItem {
	return &ListwmItem{
		Value1: nil,
		Value2: nil,
		Value3: nil,
		Value4: nil,
		Value5: nil,
		Value6: nil,
		IValue: 0,
		Prev:   nil,
		Next:   nil,
	}
}

// ListWM represents a wrapped list structure.
type ListWM struct {
	First *ListwmItem
	Last  *ListwmItem
	Count int
}

// NewListWM creates a new ListWM instance.
func NewListWM() *ListWM {
	return &ListWM{
		First: nil,
		Last:  nil,
		Count: 0,
	}
}

// DeleteAll removes all elements from the list.
func (l *ListWM) DeleteAll() {
	current := l.First
	for current != nil {
		next := current.Next
		current = next
	}
	l.First = nil
	l.Last = nil
	l.Count = 0
}

// Add adds a new item to the list.
func (l *ListWM) Add(value1, value2, value3, value4, value5, value6 interface{}, ivalue int) {
	item := NewListwmItem()

	item.Value1 = value1
	item.Value2 = value2
	item.Value3 = value3
	item.Value4 = value4
	item.Value5 = value5
	item.Value6 = value6
	item.IValue = ivalue

	if l.First == nil {
		l.First = item
	} else {
		l.Last.Next = item
		item.Prev = l.Last
	}

	l.Last = item
	l.Count++
}

// DeleteItem deletes the specified item from the list.
func (l *ListWM) DeleteItem(item *ListwmItem) {
	if item == nil || l.Count == 0 {
		return
	}

	if item == l.First {
		l.First = item.Next
	}

	if item.Prev != nil {
		item.Prev.Next = item.Next
	}

	if item == l.Last {
		l.Last = item.Prev
	}

	if item.Next != nil {
		item.Next.Prev = item.Prev
	}

	l.Count--
}

// DeleteItems deletes items with specified pointers.
func (l *ListWM) DeleteItems(value1, value2, value3, value4, value5 interface{}) {
	current := l.First
	for current != nil {
		next := current.Next

		del := true

		if value1 != nil && current.Value1 != value1 {
			del = false
		} else if value2 != nil && current.Value2 != value2 {
			del = false
		} else if value3 != nil && current.Value3 != value3 {
			del = false
		} else if value4 != nil && current.Value4 != value4 {
			del = false
		} else if value5 != nil && current.Value5 != value5 {
			del = false
		}

		if del {
			if current == l.First {
				l.First = next

				if next != nil {
					next.Prev = nil
				}
			} else if current.Prev != nil {
				current.Prev.Next = next
			}

			if current == l.Last {
				l.Last = current.Prev

				if l.Last != nil {
					l.Last.Next = nil
				}
			} else if next != nil {
				next.Prev = current.Prev
			}

			l.Count--
		}

		current = next
	}
}

// GetFirst returns the first item in the list.
func (l *ListWM) GetFirst() *ListwmItem {
	return l.First
}

// GetLast returns the last item in the list.
func (l *ListWM) GetLast() *ListwmItem {
	return l.Last
}

// GetNth returns the Nth item in the list.
func (l *ListWM) GetNth(n int) *ListwmItem {
	k := 0
	for i := l.First; i != nil; i = i.Next {
		if k == n {
			return i
		}
		k++
	}
	return nil
}

// GetCount returns the total number of items in the list.
func (l *ListWM) GetCount() int {
	return l.Count
}
