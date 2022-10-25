package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
)

var www *TheWeb
var simulation *simgo.Simulation

func main() {
	www = LoadWeb("webdata.yaml")

	sim := simgo.Simulation{}
	simulation = &sim
	frontier := NewFrontier(simulation)

	fmt.Printf("Seed list\n%s\n", frontier.Config.Seeds.String())
	fmt.Printf("URL queue\n%s\n", frontier.urlQueue.String())

	sim.ProcessReflect(Harvester, frontier)
	sim.ProcessReflect(Harvester, frontier)

	sim.RunUntil(100)
	fmt.Printf("\nURL queue\n%s\n", frontier.urlQueue.String())

	www.PrintStats()
}

func Harvester(proc simgo.Process, frontier *Frontier) {
	for {
		qUrl := frontier.GetNextToFetch()
		if qUrl == nil {
			proc.Wait(proc.Timeout(1))
			continue
		}

		fmt.Printf("[%4.0f] \u23f1  %s\n", proc.Now(), qUrl)

		r := www.Fetch(qUrl.Url)
		proc.Wait(proc.Timeout(1))
		frontier.DoneFetching(qUrl, r)
		fmt.Printf("[%4.0f] \u2714  %s\n", proc.Now(), qUrl)
	}
}
