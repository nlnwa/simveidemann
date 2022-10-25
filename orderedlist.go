package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type OrderedList struct {
	keys   [][]byte
	values [][]byte
	lock   sync.Mutex
}

func Encode(s interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(s)
	return buf.Bytes()
}

func Decode(b []byte, e interface{}) {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	_ = dec.Decode(e)
}

func (l *OrderedList) Put(key, value []byte) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.unlockedPut(key, value)
}

func (l *OrderedList) unlockedPut(key, value []byte) {
	if len(value) == 0 {
		panic("empty value is not supported")
	}

	i := sort.Search(len(l.keys), func(i int) bool { return bytes.Compare(l.keys[i], key) > 0 })

	if i > 0 && i <= len(l.keys) && bytes.Equal(l.keys[i-1], key) {
		l.values[i-1] = value
	} else {
		l.keys = append(l.keys, nil)
		l.values = append(l.values, nil)
		copy(l.keys[i+1:], l.keys[i:])
		copy(l.values[i+1:], l.values[i:])
		l.keys[i] = key
		l.values[i] = value
	}
}

func (l *OrderedList) Get(key []byte) []byte {
	i := sort.Search(len(l.keys), func(i int) bool { return bytes.Compare(l.keys[i], key) >= 0 })
	if i < len(l.keys) && bytes.Equal(l.keys[i], key) {
		return l.values[i]
	}
	return nil
}

func (l *OrderedList) Delete(key []byte) bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.unlockedDelete(key)
}

func (l *OrderedList) unlockedDelete(key []byte) bool {
	i := sort.Search(len(l.keys), func(i int) bool { return bytes.Compare(l.keys[i], key) >= 0 })
	if i < len(l.keys) && bytes.Equal(l.keys[i], key) {
		copy(l.keys[i:], l.keys[i+1:])
		copy(l.values[i:], l.values[i+1:])
		l.keys = l.keys[:len(l.keys)-1]
		l.values = l.values[:len(l.values)-1]
		return true
	}
	return false
}

func (l *OrderedList) GetIndex(idx int) []byte {
	i := l.values[idx]
	return i
}

func (l *OrderedList) Scan(from, to []byte, limit int) (keys, values [][]byte) {
	if len(l.keys) == 0 {
		return
	}

	i := 0
	if from != nil {
		i = sort.Search(len(l.keys), func(i int) bool {
			return bytes.Compare(l.keys[i], from) >= 0
		})
	}
	j := len(l.keys)
	if to != nil {
		j = sort.Search(len(l.keys), func(i int) bool {
			return bytes.Compare(l.keys[i], to) >= 0
		})
	}
	j = min(j, i+limit)

	return l.keys[i:j], l.values[i:j]
}

func (l *OrderedList) CompareAndSwap(key, previousValue, newValue []byte) ([]byte, bool) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if len(newValue) == 0 {
		panic("empty value is not supported")
	}

	v := l.Get(key)
	if bytes.Equal(v, previousValue) {
		l.unlockedPut(key, newValue)
		return v, true
	} else {
		return v, false
	}
}

func (l *OrderedList) Size() int {
	l.lock.Lock()
	defer l.lock.Unlock()

	return len(l.keys)
}

func (l *OrderedList) StringT(e interface{}) string {
	sb := strings.Builder{}
	t := reflect.TypeOf(e)
	for i, key := range l.keys {
		n := reflect.New(t)
		Decode(l.values[i], n.Interface())
		switch t.Kind() {
		case reflect.Struct:
			sb.WriteString(fmt.Sprintf("%03d %s: %s\n", i, key, n))
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			sb.WriteString(fmt.Sprintf("%03d %s: %f\n", i, key, n.Elem().Float()))
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			sb.WriteString(fmt.Sprintf("%03d %s: %d\n", i, key, n.Elem().Int()))
		default:
			panic("Unknown type " + t.Kind().String())
		}
	}
	return sb.String()
}

func (l *OrderedList) String() string {
	sb := strings.Builder{}
	for i, key := range l.keys {
		sb.WriteString(fmt.Sprintf("%03d %s: %v\n", i, key, l.values[i]))
	}
	return sb.String()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
