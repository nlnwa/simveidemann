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

func NewFrontier(sim *simgo.Simulation, configFile string) *Frontier {
	f := &Frontier{
		sim:      sim,
		Config:   LoadConfig(configFile),
		urlQueue: NewUrlQueue(sim),
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
			Seed:  &s,
		}
		f.urlQueue.Put(qurl)
	}

	keys, vals := f.Config.HostAliases.Scan(nil, nil, 1000)
	for i := 0; i < len(keys); i++ {
		var v HostAlias
		Decode(vals[i], &v)
		f.urlQueue.hostReservationService.AddHostAlias(&v)
	}

	return f
}

func (f *Frontier) GetNextToFetch() *QueuedUrl {
	host := f.urlQueue.hostReservationService.ReserveNextHost()
	if host == "" {
		return nil
	}

	from := []byte(host + " ")
	to := []byte(fmt.Sprintf("%s %05d \uffff", host, int(simulation.Now())))
	_, values := f.urlQueue.Scan(from, to, 1)

	if len(values) == 0 {
		// No URL ready for host. Release with ts for next URL in queue or delete host if no url in queue
		from := []byte(host + " ")
		to := []byte(fmt.Sprintf("%s \xff", host))
		_, values = f.urlQueue.Scan(from, to, 1)
		if len(values) > 0 {
			var q QueuedUrl
			Decode(values[0], &q)
			f.urlQueue.hostReservationService.ReleaseHost(host, q.Ts)
		} else {
			fmt.Printf("[%4.0f]    Host '%s' deleted\n", simulation.Now(), host)
			f.urlQueue.hostReservationService.DeleteHost(host)
		}
		return nil
	}

	var q QueuedUrl
	Decode(values[0], &q)
	qUrl := &q

	f.urlQueue.SetBusy(qUrl)
	return qUrl
}

func (f *Frontier) DoneFetching(qUrl *QueuedUrl, response *WebResponse) {
	politeness := 0

	var opts []Opt
	if response.Status == 1500 {
		opts = append(opts, FailUnsetBusyHost)
	}
	defer f.urlQueue.hostReservationService.ReleaseHost(NormalizeHost(qUrl.Url), int(simulation.Now())+politeness, opts...)

	seed := f.Config.Seeds.GetBestSeedForUrl(qUrl.Url)
	profiles := f.findProfiles(seed, qUrl)
	if len(profiles) > 0 {
		ts := math.MaxInt32
		for _, profile := range profiles {
			t := f.calcDelay(qUrl, profile)
			if ts > t {
				ts = t
			}
		}
		f.urlQueue.SetIdle(qUrl, ts)
		for _, o := range response.Outlinks {
			f.handleOutlink(qUrl, o)
		}
	} else {
		f.urlQueue.Delete(qUrl)
	}
}

func (f *Frontier) handleOutlink(qUrl *QueuedUrl, outlink string) {
	u, _ := whatwgUrl.Parse(outlink)
	ol := &QueuedUrl{
		Host:  u.Hostname(),
		Ts:    qUrl.Ts,
		Url:   outlink,
		Level: qUrl.Level + 1,
		Seed:  qUrl.Seed,
	}

	// Maybe there is a Seed with a more specific match
	seed := f.Config.Seeds.GetBestSeedForUrl(ol.Url)
	if seed != nil {
		ol.Seed = seed
	}

	if len(f.findProfiles(ol.Seed, ol)) > 0 {
		f.urlQueue.Put(ol)
	}
}

func (f *Frontier) findProfiles(seed *Seed, qUrl *QueuedUrl) []*Profile {
	var pr []*Profile
	if seed != nil {
		for _, p := range seed.Profiles {
			if f.scopeCheck(p, qUrl) {
				pr = append(pr, p)
			}
		}
	}
	return pr
}

func (f *Frontier) calcDelay(qUrl *QueuedUrl, profile *Profile) int {
	return qUrl.LastFetch + profile.Freq
}

func (f *Frontier) scopeCheck(profile *Profile, qUrl *QueuedUrl) bool {
	return qUrl.Level <= profile.Scope
}
