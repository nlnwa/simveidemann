package main

import (
	"fmt"
)

type UrlQueue struct {
	OrderedList
	valueType interface{}
}

func (uq *UrlQueue) Put(key []byte, value *QueuedUrl) {
	uq.OrderedList.Put(key, Encode(value))
}

func (uq *UrlQueue) Get(key []byte) []byte {
	return uq.OrderedList.Get(key)
}

func (uq *UrlQueue) Delete(key []byte) bool {
	return uq.OrderedList.Delete(key)
}

func (uq *UrlQueue) String() string {
	return uq.OrderedList.StringT(uq.valueType)
}

type QueuedUrl struct {
	Busy      bool
	Host      string
	Ts        int
	Url       string
	Level     int
	LastFetch int
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
	return []byte(fmt.Sprintf("%s %s%05d %s", NormalizedHost(u.Url), busy, u.Ts, u.Url))
}
