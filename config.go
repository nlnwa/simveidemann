package main

import (
	"bytes"
	"fmt"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type Config struct {
	Profiles map[string]*Profile
	Seeds    *Seeds
}

type Seeds struct {
	OrderedList
	valueType interface{}
}

func (sl *Seeds) String() string {
	return sl.OrderedList.StringT(sl.valueType)
}

type Seed struct {
	Url      string
	Profiles []*Profile
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

	keys, values := sl.Scan([]byte(from), []byte(to), 1000)

	for i := len(keys) - 1; i >= 0; i-- {
		if bytes.Equal(oto, keys[i]) {
			var s Seed
			Decode(values[i], &s)
			return &s
		} else if bytes.HasPrefix(oto, keys[i]) {
			k := bytes.TrimRight(keys[i], "/")
			if oto[len(k)] == '/' {
				var s Seed
				Decode(values[i], &s)
				return &s
			}
		}
	}

	return nil
}

func LoadConfig(configFile string) *Config {
	c := &Config{
		Profiles: make(map[string]*Profile),
		Seeds: &Seeds{
			valueType: Seed{},
		},
	}

	f1, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f1.Close() }()

	d := yaml.NewDecoder(f1)

	type SeedConfig struct {
		Url      string
		Profiles []string
	}

	type Config struct {
		Profiles []*Profile
		Seeds    []*SeedConfig
	}

	in := Config{}

	err = d.Decode(&in)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	for _, p := range in.Profiles {
		if p.Name == "" {
			continue
		}
		c.Profiles[p.Name] = p
	}

	for _, s := range in.Seeds {
		if s.Url == "" {
			continue
		}
		seed := &Seed{
			Url: s.Url,
		}
		for _, p := range s.Profiles {
			seed.Profiles = append(seed.Profiles, c.Profiles[p])
		}
		c.Seeds.Put(seed.Key(), Encode(seed))
	}

	return c
}
