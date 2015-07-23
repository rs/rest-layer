package mem

import (
	"time"

	"github.com/rs/rest-layer"
)

// sortableItems is an item slice implementing sort.Interface
type sortableItems struct {
	sort  []string
	items []*rest.Item
}

func (s sortableItems) Len() int {
	return len(s.items)
}

func (s sortableItems) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (s sortableItems) Less(i, j int) bool {
	for _, exp := range s.sort {
		var field1 interface{}
		var field2 interface{}
		if exp[0] == '-' {
			field1 = s.items[j].GetField(exp[1:])
			field2 = s.items[i].GetField(exp[1:])
		} else {
			field1 = s.items[i].GetField(exp)
			field2 = s.items[j].GetField(exp)
		}
		if field1 == field2 {
			continue
		}
		switch t := field1.(type) {
		case int:
			return t < field2.(int)
		case int8:
			return t < field2.(int8)
		case int16:
			return t < field2.(int16)
		case int32:
			return t < field2.(int32)
		case int64:
			return t < field2.(int64)
		case uint:
			return t < field2.(uint)
		case uint8:
			return t < field2.(uint8)
		case uint16:
			return t < field2.(uint16)
		case uint32:
			return t < field2.(uint32)
		case uint64:
			return t < field2.(uint64)
		case float32:
			return t < field2.(float32)
		case float64:
			return t < field2.(float64)
		case string:
			return t < field2.(string)
		case bool:
			return t
		case time.Time:
			return t.Before(field2.(time.Time))
		}
	}
	return false
}
