package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
)

var www *TheWeb

func main() {
	www = &TheWeb{resources: make(map[string]*WebResource)}
	loadWeb(www, "webdata.yaml")

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

type TheWeb struct {
	resources map[string]*WebResource
}

type WebResource struct {
	Url      string
	Outlinks []string
	Status   int
}

func (w *TheWeb) Fetch(u string) *WebResource {
	var v *WebResource
	var ok bool
	if v, ok = w.resources[u]; ok {
		if v.Status == 0 {
			v.Status = 200
		}
	} else {
		v = &WebResource{
			Url:      u,
			Outlinks: nil,
			Status:   404,
		}
	}
	return v
}
