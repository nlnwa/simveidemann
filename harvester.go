package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
)

type Harvester struct {
	busy int
}

func NewHarvester(sim *simgo.Simulation, frontier *Frontier) *Harvester {
	h := &Harvester{}
	sim.ProcessReflect(h.Run, frontier)
	return h
}

func (h *Harvester) Run(proc simgo.Process, frontier *Frontier) {
	for {
		qUrl := frontier.GetNextToFetch()
		if qUrl == nil {
			proc.Wait(proc.Timeout(1))
			continue
		}

		fmt.Printf("[%4.0f] \u23f1  %s\n", proc.Now(), qUrl)
		start := int(proc.Now())
		r := www.Fetch(qUrl.Url)
		proc.Wait(proc.Timeout(1))
		frontier.DoneFetching(qUrl, r)
		h.busy = h.busy + int(proc.Now()) - start
		fmt.Printf("[%4.0f] \u2714  %4d %s\n", proc.Now(), r.Status, qUrl)
	}
}
