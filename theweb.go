package main

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

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
	if v, ok = w.resources[u]; !ok {
		v = &WebResource{
			Url:      u,
			Outlinks: nil,
			Status:   404,
		}
	}
	return v
}

func LoadWeb(file string) *TheWeb {
	www = &TheWeb{resources: make(map[string]*WebResource)}

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	d := yaml.NewDecoder(f)
	r := make([]*WebResource, 0)
	err = d.Decode(&r)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	for _, w := range r {
		if w.Url == "" {
			continue
		}
		if w.Status == 0 {
			w.Status = 200
		}
		www.resources[w.Url] = w
	}

	return www
}
