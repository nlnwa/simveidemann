package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
)

type UrlQueue struct {
	OrderedList
	valueType              interface{}
	knownUrls              *OrderedList
	hostReservationService *HostReservationService
	sim                    *simgo.Simulation
}

func NewUrlQueue(sim *simgo.Simulation) *UrlQueue {
	return &UrlQueue{
		valueType:              QueuedUrl{},
		knownUrls:              &OrderedList{},
		hostReservationService: NewHostReservationService(sim),
		sim:                    sim,
	}
}

func (uq *UrlQueue) Put(qUrl *QueuedUrl) {
	key := uq.urlKey(qUrl)
	if len(uq.knownUrls.Get([]byte(qUrl.Url))) == 0 {
		uq.knownUrls.Put([]byte(qUrl.Url), nullArray)
		uq.OrderedList.Put(key, Encode(qUrl))
		uq.hostReservationService.AddHost(NormalizeHost(qUrl.Url))
	}
}

func (uq *UrlQueue) Delete(qUrl *QueuedUrl) bool {
	key := uq.urlKey(qUrl)
	uq.knownUrls.Delete([]byte(qUrl.Url))
	return uq.OrderedList.Delete(key)
}

func (uq *UrlQueue) SetBusy(qUrl *QueuedUrl) {
	now := int(uq.sim.Now())
	idleKey := uq.urlKey(qUrl)
	qUrl.Busy = true
	qUrl.LastFetch = now
	qUrl.Ts = now + busyTimeout
	busyKey := uq.urlKey(qUrl)
	uq.OrderedList.Put(busyKey, Encode(qUrl))
	uq.OrderedList.Delete(idleKey)
}

func (uq *UrlQueue) SetIdle(qUrl *QueuedUrl, newTS int) {
	busyKey := uq.urlKey(qUrl)
	qUrl.Busy = false
	qUrl.Ts = newTS
	idleKey := uq.urlKey(qUrl)
	uq.OrderedList.Put(idleKey, Encode(qUrl))
	uq.OrderedList.Delete(busyKey)
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
	Seed      *Seed
}

func (u *QueuedUrl) String() string {
	return fmt.Sprintf("%s, hops: %d", u.Url, u.Level)
}

func (uq *UrlQueue) urlKey(u *QueuedUrl) []byte {
	if u == nil {
		return []byte{}
	}
	busy := ""
	if u.Busy {
		busy = "b"
	}
	return []byte(fmt.Sprintf("%s %s%05d %s", NormalizeHost(u.Url), busy, u.Ts, u.Url))
}
