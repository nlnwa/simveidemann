package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"
	"math"
)

type Frontier struct {
	sim                    *simgo.Simulation
	Config                 *Config
	urlQueue               *UrlQueue
	hostReservationService *HostReservationService
}

func NewFrontier(sim *simgo.Simulation, configFile string) *Frontier {
	f := &Frontier{
		sim:    sim,
		Config: LoadConfig(configFile),
		urlQueue: &UrlQueue{
			valueType: QueuedUrl{},
		},
		hostReservationService: NewHostReservationService(),
	}
	for i := 0; i < f.Config.Seeds.Size(); i++ {
		s := Seed{}
		Decode(f.Config.Seeds.GetIndex(i), &s)
		u, _ := whatwgUrl.Parse(s.Url)
		qurl := &QueuedUrl{
			Host:  u.Hostname(),
			Ts:    0,
			Url:   s.Url,
			Level: 0,
		}
		f.urlQueue.Put(qurl.Key(), qurl)
		f.hostReservationService.AddHost(NormalizedHost(s.Url))
	}

	return f
}

func (f *Frontier) GetNextToFetch() *QueuedUrl {
	host := f.hostReservationService.ReserveNextHost()
	if host == "" {
		return nil
	}

	from := []byte(host + " ")
	to := []byte(fmt.Sprintf("%s %05d \uffff", host, int(simulation.Now())))
	keys, values := f.urlQueue.Scan(from, to, 1)

	if len(keys) == 0 {
		from := []byte(host + " ")
		to := []byte(fmt.Sprintf("%s \uffff", host))
		keys, values = f.urlQueue.Scan(from, to, 1)
		if len(keys) > 0 {
			var q QueuedUrl
			Decode(values[0], &q)
			f.hostReservationService.ReleaseHost(host, q.Ts)
		} else {
			f.hostReservationService.ReleaseHost(host, 2)
		}
		return nil
	}

	var q QueuedUrl
	Decode(values[0], &q)
	qUrl := &q

	f.urlQueue.Delete(qUrl.Key())
	qUrl.Busy = true
	qUrl.LastFetch = int(f.sim.Now())
	f.urlQueue.Put(qUrl.Key(), qUrl)
	return qUrl
}

func (f *Frontier) DoneFetching(qUrl *QueuedUrl, response *WebResource) {
	politeness := 0
	defer f.hostReservationService.ReleaseHost(NormalizedHost(qUrl.Url), int(simulation.Now())+politeness)
	f.urlQueue.Delete(qUrl.Key())
	qUrl.Busy = false
	profiles := f.findProfiles(qUrl)
	if len(profiles) > 0 {
		qUrl.Ts = math.MaxInt32
		for _, profile := range profiles {
			f.calcDelay(qUrl, profile)
		}
		f.urlQueue.Put(qUrl.Key(), qUrl)
		for _, o := range response.Outlinks {
			f.handleOutlink(qUrl, o)
		}
	}
}

func (f *Frontier) handleOutlink(qUrl *QueuedUrl, outlink string) {
	u, _ := whatwgUrl.Parse(outlink)
	ol := &QueuedUrl{
		Host:  u.Hostname(),
		Ts:    qUrl.Ts,
		Url:   outlink,
		Level: qUrl.Level + 1,
	}
	if len(f.findProfiles(ol)) > 0 {
		f.urlQueue.Put(ol.Key(), ol)
	}
}

func (f *Frontier) findProfiles(qUrl *QueuedUrl) []*Profile {
	var pr []*Profile
	seed := f.Config.Seeds.GetBestSeedForUrl(qUrl.Url)
	if seed != nil {
		for _, p := range seed.Profiles {
			if f.scopeCheck(p, qUrl) {
				pr = append(pr, p)
			}
		}
	}
	return pr
}

func (f *Frontier) calcDelay(qUrl *QueuedUrl, profile *Profile) {
	ts := int(qUrl.LastFetch) + profile.Freq
	if qUrl.Ts > ts {
		qUrl.Ts = ts
	}
}

func (f *Frontier) scopeCheck(profile *Profile, qUrl *QueuedUrl) bool {
	return qUrl.Level <= profile.Scope
}
