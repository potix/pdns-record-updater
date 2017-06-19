package helper

import (
	"sort"
)

type Sorter struct {
	object   interface{}
	GetLen   func(object interface{}) (int)
	Get      func(object interface{}, i int) (interface{})
	Put      func(object interface{}, i int, a interface{})
	Compare  func(a, b interface{}) bool
	list     []interface{}
}

func (s Sorter) Len() int {
	return len(s.list)
}

func (s Sorter) Swap(i, j int) {
	s.list[i], s.list[j] = s.list[j], s.list[i]
}

func (s Sorter) Less(i, j int) bool {
    return s.Compare(s.list[i], s.list[j])
}

func (s *Sorter) updateList() {
	currentListLen := s.GetLen(s.object)
	if s.list == nil || len(s.list) != currentListLen {
		s.list = make([]interface{}, 0, currentListLen)
		for i := 0; i < currentListLen; i++ {
			s.list = append(s.list, s.Get(s.object, i))
		}
	}
}

func (s *Sorter) feedback() {
	if s.Put != nil {
		for i := 0; i < len(s.list); i++ {
			s.Put(s.object, i, s.list[i])
		}
	}
}

func (s *Sorter) Sort() ([]interface{}) {
	s.updateList()
	sort.Sort(s)
	s.feedback()
	return s.list
}

func (s *Sorter) ReverseSort() ([]interface{}){
	s.updateList()
	sort.Sort(sort.Reverse(s))
	s.feedback()
	return s.list
}

// usage
//
//	sorter := &Sorter {
//		Object: myList,
//		GetLen : func(object interface{}) (int) {
//			return len(object.([]*my))
//		},
//		Get : func(object interface{}, i int) (interface{}) {
//			return object.([]*my)[i]
//		},
//		Put : func(object interface{}, i int, a interface{}) {
//			object.([]*my)[i] = a.(*my)
//		},
//		Compare : func(a, b interface{}) bool {
//			return a.(*my).member < b.(*my).member
//		},
//	}
//	sorter.Sort()
//	sorter.ReverseSort()
