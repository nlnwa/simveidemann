package main

import (
	"bytes"
	"fmt"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"
)

type Config struct {
	Profiles map[string]*Profile
	Seeds    *Seeds
}

func NewConfig() *Config {
	c := &Config{
		Profiles: make(map[string]*Profile),
		Seeds: &Seeds{
			valueType: Seed{},
		},
	}
	return c
}

type Seeds struct {
	OrderedList
	valueType interface{}
}

func (sl *Seeds) String() string {
	return sl.OrderedList.String(sl.valueType)
}

type Seed struct {
	Url      string
	Profiles []*Profile
}

type SeedConfig struct {
	Url      string
	Profiles []string
}

func (s *Seed) String() string {
	return fmt.Sprintf("%s %s", s.Url, s.Profiles)
}

func (s *Seed) Key() []byte {
	u, _ := whatwgUrl.Parse(s.Url)
	return []byte(u.Hostname() + u.Pathname())
}

type Profile struct {
	Name  string
	Scope int
	Freq  int
}

func (p Profile) String() string {
	return fmt.Sprintf("(%s: scope: %d, freq: %d)", p.Name, p.Scope, p.Freq)
}

func (sl *Seeds) GetBestSeedForUrl(u string) *Seed {
	pu, _ := whatwgUrl.Parse(u)
	from := pu.Hostname()
	to := pu.Hostname() + pu.Pathname()
	from += "/"
	oto := []byte(to)
	to = to + "\000"

	candidates := sl.Scan([]byte(from), []byte(to), 1000)

	for i := len(candidates) - 1; i >= 0; i-- {
		cand := candidates[i]

		if bytes.Equal(oto, cand.key) {
			var s Seed
			Decode(cand.value, &s)
			return &s
		} else if bytes.HasPrefix(oto, cand.key) {
			k := bytes.TrimRight(cand.key, "/")
			if oto[len(k)] == '/' {
				var s Seed
				Decode(cand.value, &s)
				return &s
			}
		}
	}

	return nil
}
