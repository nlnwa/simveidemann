package main

import (
	"fmt"
	"github.com/fschuetz04/simgo"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"
	"math"
)

type Frontier struct {
	sim      *simgo.Simulation
	Config   *Config
	urlQueue *UrlQueue
}

func NewFrontier(sim *simgo.Simulation) *Frontier {
	f := &Frontier{
		sim:    sim,
		Config: LoadConfig("config.yaml"),
		urlQueue: &UrlQueue{
			valueType: QueuedUrl{},
		},
	}
	for i := 0; i < len(f.Config.Seeds.OrderedList); i++ {
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
	}
	return f
}

func (f *Frontier) GetNextToFetch() *QueuedUrl {
	var qUrl *QueuedUrl
	for i := 0; i < len(f.urlQueue.OrderedList); i++ {
		d := f.urlQueue.GetIndex(i)
		var q QueuedUrl
		Decode(d, &q)
		qUrl = &q
		now := fmt.Sprintf("%05.0f", f.sim.Now())
		if !qUrl.Busy && string(qUrl.Key()) <= now {
			break
		}
		qUrl = nil
	}
	if qUrl == nil {
		return nil
	}

	f.urlQueue.Delete(qUrl.Key())
	qUrl.Busy = true
	f.urlQueue.Put(qUrl.Key(), qUrl)
	return qUrl
}

func (f *Frontier) DoneFetching(qUrl *QueuedUrl, response *WebResource) {
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
	for _, p := range seed.Profiles {
		if f.scopeCheck(p, qUrl) {
			pr = append(pr, p)
		}
	}
	return pr
}

func (f *Frontier) calcDelay(qUrl *QueuedUrl, profile *Profile) {
	ts := int(f.sim.Now()) + profile.Freq
	if qUrl.Ts > ts {
		qUrl.Ts = ts
	}
}

func (f *Frontier) scopeCheck(profile *Profile, qUrl *QueuedUrl) bool {
	return qUrl.Level <= profile.Scope
}
