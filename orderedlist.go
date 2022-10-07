package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type listItem struct {
	key   []byte
	value []byte
}

type OrderedList []*listItem

func Encode(s interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(s)
	return buf.Bytes()
}

func Decode(b []byte, e interface{}) {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	dec.Decode(e)
}

func (l *OrderedList) Put(key, value []byte) {
	li := &listItem{
		key:   key,
		value: value,
	}

	i := sort.Search(len(*l), func(i int) bool { return bytes.Compare((*l)[i].key, li.key) > 0 })
	*l = append(*l, nil)
	copy((*l)[i+1:], (*l)[i:])
	(*l)[i] = li
}

func (l *OrderedList) Get(key []byte) []byte {
	i := sort.Search(len(*l), func(i int) bool { return bytes.Compare((*l)[i].key, key) >= 0 })
	if i < len(*l) && bytes.Equal((*l)[i].key, key) {
		return (*l)[i].value
	}
	return nil
}

func (l *OrderedList) Delete(key []byte) bool {
	i := sort.Search(len(*l), func(i int) bool { return bytes.Compare((*l)[i].key, key) >= 0 })
	if i < len(*l) && bytes.Equal((*l)[i].key, key) {
		copy((*l)[i:], (*l)[i+1:])
		*l = (*l)[:len(*l)-1]
		return true
	}
	return false
}

func (l *OrderedList) GetIndex(idx int) []byte {
	i := (*l)[idx]
	return i.value
}

func (l *OrderedList) Scan(from, to []byte, limit int) OrderedList {
	i := 0
	if from != nil {
		i = sort.Search(len(*l), func(i int) bool {
			return bytes.Compare((*l)[i].key, from) >= 0
		})
	}
	j := len(*l) - 1
	if to != nil {
		j = sort.Search(len(*l), func(i int) bool {
			return bytes.Compare((*l)[i].key, to) >= 0
		})
	}
	j = min(j, i+limit)

	return (*l)[i:j]
}

func (l *OrderedList) String(e interface{}) string {
	sb := strings.Builder{}
	t := reflect.TypeOf(e)
	for i, li := range *l {
		n := reflect.New(t)
		Decode(li.value, n.Interface())
		sb.WriteString(fmt.Sprintf("%03d %s: %s\n", i, li.key, n))
	}
	return sb.String()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
