package main

import (
	"fmt"
)

type UrlQueue struct {
	OrderedList
	valueType interface{}
	urlIndex  OrderedList
}

func (uq *UrlQueue) Put(key []byte, value *QueuedUrl) {
	uq.OrderedList.Put(key, Encode(value))
	busy := ""
	uq.urlIndex.Put([]byte(value.Url), Encode(&indexItem{value.Url, fmt.Sprintf("%s%05d", busy, value.Ts)}))
}

func (uq *UrlQueue) Get(key []byte) []byte {
	return uq.OrderedList.Get(key)
}

func (uq *UrlQueue) Delete(key []byte) bool {
	return uq.OrderedList.Delete(key)
}

func (uq *UrlQueue) Scan(from, to []byte, limit int) OrderedList {
	return uq.OrderedList.Scan(from, to, limit)
}

func (uq *UrlQueue) String() string {
	return uq.OrderedList.String(uq.valueType)
}

type QueuedUrl struct {
	Busy  bool
	Host  string
	Ts    int
	Url   string
	Level int
}

func (u *QueuedUrl) String() string {
	return fmt.Sprintf("%s, hops: %d", u.Url, u.Level)
}

func (u *QueuedUrl) Key() []byte {
	if u == nil {
		return []byte{}
	}
	busy := ""
	if u.Busy {
		busy = "b"
	}
	return []byte(fmt.Sprintf("%s%05d", busy, u.Ts))
}

type indexItem struct {
	K string
	V string
}

func (ii *indexItem) Key() []byte {
	return []byte(ii.K)
}

func (ii *indexItem) String() string {
	return fmt.Sprintf("{Url: %s, Ts: %s}", ii.K, ii.V)
}
