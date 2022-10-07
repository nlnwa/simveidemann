package main

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func loadWeb(www *TheWeb, file string) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	d := yaml.NewDecoder(f)

	for {
		w := WebResource{}
		err = d.Decode(&w)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if w.Url == "" {
			continue
		}
		www.resources[w.Url] = &w
	}
}

func (c *Config) loadConfig(profilesFile, seedsFile string) {
	f1, err := os.Open(profilesFile)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f1.Close() }()

	d := yaml.NewDecoder(f1)

	for {
		p := Profile{}
		err = d.Decode(&p)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if p.Name == "" {
			continue
		}
		c.Profiles[p.Name] = &p
	}

	f2, err := os.Open(seedsFile)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f2.Close() }()

	d = yaml.NewDecoder(f2)

	for {
		s := SeedConfig{}
		err = d.Decode(&s)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
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
}
