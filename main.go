package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
	flag "github.com/spf13/pflag"
)

var www *TheWeb
var simulation *simgo.Simulation

func main() {
	var webFile string
	var configFile string
	var numHarvesters int
	var runTime int
	flag.StringVarP(&webFile, "webdata", "w", "webdata.yaml", "File with web data")
	flag.StringVarP(&configFile, "config", "c", "config.yaml", "File with crawler config (seeds, profiles)")
	flag.IntVarP(&numHarvesters, "harvesters", "H", 2, "Number of harvesters")
	flag.IntVarP(&runTime, "time", "t", 100, "Length of simulation")
	flag.Parse()

	www = LoadWeb(webFile)

	sim := simgo.Simulation{}
	simulation = &sim
	frontier := NewFrontier(simulation, configFile)

	fmt.Printf("Seed list\n%s\n", frontier.Config.Seeds.String())
	fmt.Printf("URL queue\n%s\n", frontier.urlQueue.String())
	fmt.Printf("Host queue\n%s\n", frontier.hostReservationService.HostQueue.String())
	//fmt.Printf("Host alias\n%s\n", frontier.hostReservationService.HostAlias.String())

	var harvesters []*Harvester
	for i := 0; i < numHarvesters; i++ {
		harvesters = append(harvesters, NewHarvester(simulation, frontier))
	}

	sim.RunUntil(float64(runTime))
	fmt.Printf("\nURL queue\n%s\n", frontier.urlQueue.String())

	www.PrintStats()
	fmt.Println()

	for i, h := range harvesters {
		fmt.Printf("Harvester #%d: Busy: %v, Idle: %v, Load: %.1f%%\n", i, h.busy, runTime-h.busy, (float64(h.busy)/float64(runTime))*100)
	}
}

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
		fmt.Printf("[%4.0f] \u2714  %s\n", proc.Now(), qUrl)
	}
}
