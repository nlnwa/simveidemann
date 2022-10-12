package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
)

var www *TheWeb

func main() {
	www = LoadWeb("webdata.yaml")

	sim := simgo.Simulation{}
	frontier := NewFrontier(&sim)

	fmt.Printf("Seed list\n%s\n", frontier.Config.Seeds.String())
	fmt.Printf("URL queue\n%s\n", frontier.urlQueue.String())

	sim.ProcessReflect(Harvester, frontier)
	sim.ProcessReflect(Harvester, frontier)

	sim.RunUntil(100)
	fmt.Printf("\nURL queue\n%s\n", frontier.urlQueue.String())
	fmt.Printf("URL index\n%s\n", frontier.urlQueue.urlIndex.String(indexItem{}))
}

func Harvester(proc simgo.Process, frontier *Frontier) {
	for {
		qUrl := frontier.GetNextToFetch()
		if qUrl == nil {
			proc.Wait(proc.Timeout(0.1))
			continue
		}

		fmt.Printf("[%4.0f] \u23f1  %s\n", proc.Now(), qUrl)

		proc.Wait(proc.Timeout(10))
		r := www.Fetch(qUrl.Url)
		frontier.DoneFetching(qUrl, r)
		fmt.Printf("[%4.0f] \u2714  %s\n", proc.Now(), qUrl)
	}
}
