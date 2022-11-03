package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
	"github.com/nlnwa/whatwg-url/canonicalizer"
	"github.com/nlnwa/whatwg-url/url"
	"strings"
)

const busyTimeout = 20

var nullArray = []byte{0}

type HostReservationService struct {
	HostQueue *OrderedList
	Hosts     *OrderedList
	HostAlias *OrderedList
}

func NewHostReservationService(sim *simgo.Simulation) *HostReservationService {
	hrs := &HostReservationService{
		HostQueue: &OrderedList{},
		Hosts:     &OrderedList{},
		HostAlias: &OrderedList{},
	}

	sim.Process(func(proc simgo.Process) {
		for {
			proc.Wait(proc.Timeout(10))

			keys, values := hrs.HostAlias.Scan(nil, nil, 1000)
			for i := 0; i < len(keys); i++ {
				if !strings.Contains(string(keys[i]), ".") {
					var value HostAlias
					Decode(values[i], &value)
					if value.BusyTS != 0 && int(sim.Now()) > value.BusyTS {
						hrs.releaseHostAlias(keys[i])
					}
				}
			}

			keys, values = hrs.Hosts.Scan(nil, nil, 1000)
			for i := 0; i < len(keys); i++ {
				var value int
				Decode(values[i], &value)
				if value != 0 && int(sim.Now()) > value {
					hrs.ReleaseHost(string(keys[i]), int(sim.Now()))
				}
			}
		}
	})

	return hrs
}

func (h *HostReservationService) ReserveNextHost() string {
	to := fmt.Sprintf("%05d \uffff", int(simulation.Now()))
	keys, _ := h.HostQueue.Scan(nil, []byte(to), 100)
	for _, key := range keys {
		host := key[6:]
		ts := int(simulation.Now() + busyTimeout)

		aliasName := h.HostAlias.Get(host)
		if aliasName != nil {
			if !h.reserveHostAlias(aliasName, ts) {
				continue
			}
		}

		if _, ok := h.Hosts.CompareAndSwap(host, nullArray, Encode(ts)); !ok {
			continue
		}
		h.HostQueue.Delete(key)

		return string(host)
	}

	return ""
}

func (h *HostReservationService) ReleaseHost(host string, nextTs int, opts ...Opt) {
	if host == "" {
		panic("Cannot release empty host")
	}

	h.HostQueue.Put([]byte(fmt.Sprintf("%05d %s", nextTs, host)), nullArray)

	aliasName := h.HostAlias.Get([]byte(host))
	if aliasName != nil {
		h.releaseHostAlias(aliasName)
	}

	if FailUnsetBusyHost.IsIn(opts) {
		return
	}

	p := h.Hosts.Get([]byte(host))
	h.Hosts.CompareAndSwap([]byte(host), p, nullArray)
}

func (h *HostReservationService) AddHost(host string) {
	if _, ok := h.Hosts.CompareAndSwap([]byte(host), nil, nullArray); ok {
		h.HostQueue.Put([]byte(fmt.Sprintf("%05d %s", 0, host)), nullArray)
	}
}

func (h *HostReservationService) AddHostAlias(ha *HostAlias) {
	if _, ok := h.HostAlias.CompareAndSwap([]byte(ha.Name), nil, Encode(ha)); ok {
		for _, host := range ha.Hosts {
			h.HostAlias.Put([]byte(host), []byte(ha.Name))
		}
	}
}

func (h *HostReservationService) reserveHostAlias(aliasName []byte, ts int) bool {
	aliasBytes := h.HostAlias.Get(aliasName)
	if aliasBytes == nil {
		return false
	}
	alias := HostAlias{}
	Decode(aliasBytes, &alias)
	if alias.BusyTS > 0 {
		return false
	}

	alias.BusyTS = ts
	newAlias := Encode(alias)
	if _, ok := h.HostAlias.CompareAndSwap(aliasName, aliasBytes, newAlias); ok {
		return true
	}
	return false
}

func (h *HostReservationService) releaseHostAlias(aliasName []byte) {
	aliasBytes := h.HostAlias.Get(aliasName)
	if aliasBytes == nil {
		return
	}
	alias := HostAlias{}
	Decode(aliasBytes, &alias)
	if alias.BusyTS == 0 {
		return
	}

	alias.BusyTS = 0
	newAlias := Encode(alias)
	h.HostAlias.CompareAndSwap(aliasName, aliasBytes, newAlias)
}

func NormalizeHost(u string) string {
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
