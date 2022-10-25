package main

import (
	"fmt"
	"github.com/nlnwa/whatwg-url/canonicalizer"
	"github.com/nlnwa/whatwg-url/url"
	"strings"
)

const busyTimeout = 20

var nullArray = []byte{0}

type HostReservationService struct {
	HostQueue *OrderedList
	BusyHost  *OrderedList
	HostAlias *OrderedList
	BusyAlias *OrderedList
}

func NewHostReservationService() *HostReservationService {
	return &HostReservationService{
		HostQueue: &OrderedList{},
		BusyHost:  &OrderedList{},
	}
}

func (h *HostReservationService) ReserveNextHost() string {
	to := fmt.Sprintf("%05d \uffff", int(simulation.Now()))
	keys, _ := h.HostQueue.Scan(nil, []byte(to), 1)
	if len(keys) > 0 {
		ts := Encode(int(simulation.Now() + busyTimeout))
		host := keys[0][6:]
		k := keys[0]

		v, _ := h.BusyHost.CompareAndSwap(host, nullArray, ts)
		var prev int
		Decode(v, &prev)
		h.HostQueue.Delete(k)

		return string(host)
	}

	return ""
}

func (h *HostReservationService) ReleaseHost(host string, nextTs int) {
	if host == "" {
		panic("Cannot release empty host")
	}

	p := h.BusyHost.Get([]byte(host))
	if _, ok := h.BusyHost.CompareAndSwap([]byte(host), p, nullArray); ok {
		h.HostQueue.Put([]byte(fmt.Sprintf("%05d %s", nextTs, host)), nullArray)
	}
}

func (h *HostReservationService) AddHost(host string) {
	if _, ok := h.BusyHost.CompareAndSwap([]byte(host), nil, nullArray); ok {
		h.HostQueue.Put([]byte(fmt.Sprintf("%05d %s", 0, host)), nullArray)
	}
}

func NormalizedHost(u string) string {
	c := canonicalizer.New(url.WithPostParseHostFunc(func(url *url.Url, host string) string {
		host = strings.TrimPrefix(host, "www.")
		return host
	}))
	pu, err := c.Parse(u)
	if err != nil {
		panic(err)
	}
	return pu.Hostname()
}
