package raft

// Value is a wrapper for value type in LinkedList.
type Value string

// LinkedList is a raft linked list implementation.
type LinkedList interface {
	Add(v Value)
	GetLast() *Value
	String() string
}

type SimpleLL struct {
	first *valueNode
	last  *valueNode
}

func NewSimpleLL() LinkedList {
	return &SimpleLL{}
}

type valueNode struct {
	value Value
	next  *valueNode
}

func (s *SimpleLL) Add(v Value) {
	vn := &valueNode{
		value: v,
	}
	if s.first == nil {
		s.first = vn
	}
	if s.last == nil {
		s.last = vn
	} else {
		s.last.next = vn
		s.last = vn
	}

}

func (s *SimpleLL) GetLast() *Value {
	if s.last == nil {
		return nil
	}
	return &s.last.value
}

func (s *SimpleLL) String() string {
	str := ""
	if s.first == nil {
		return str
	}
	curr := s.first
	for curr != nil {
		str += string(curr.value) + ","
		curr = curr.next
	}
	return str
}
