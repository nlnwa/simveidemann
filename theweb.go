package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"sort"
	"strings"
)

type TheWeb struct {
	resources map[string]*WebResource
	stats     map[string][]int
}

type WebResource struct {
	Url       string
	Responses []*WebResponse
	callCount int
}

type WebResponse struct {
	Outlinks []string
	Status   int
	Count    int
}

func (w *TheWeb) Fetch(u string) *WebResponse {
	var v *WebResource
	var ok bool
	if v, ok = w.resources[u]; !ok || len(v.Responses) == 0 {
		v = &WebResource{
			Url: u,
			Responses: []*WebResponse{{
				Outlinks: nil,
				Status:   404,
			}},
		}
	}
	v.callCount++
	var r *WebResponse
	c := 0
	for _, r = range v.Responses {
		c += r.Count
		if v.callCount <= c {
			break
		}
	}
	w.stats[u] = append(w.stats[u], int(simulation.Now()))
	return r
}

func LoadWeb(file string) *TheWeb {
	www = &TheWeb{
		resources: make(map[string]*WebResource),
		stats:     make(map[string][]int),
	}

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
		for _, r := range w.Responses {
			if r.Status == 0 {
				r.Status = 200
			}
			if r.Count == 0 {
				r.Count = 1
			}
		}
		www.resources[w.Url] = w
	}

	return www
}

func (w *TheWeb) PrintStats() {
	maxLen := 0
	keys := make([]string, 0, len(w.stats))
	for k, v := range w.stats {
		keys = append(keys, k)
		m := v[len(v)-1]
		if m > maxLen {
			maxLen = m
		}
	}
	sort.Strings(keys)

	b1 := strings.Builder{}
	b2 := strings.Builder{}
	b1.WriteString("0")
	b2.WriteString("\u251c")

	length := ((maxLen + 9) / 10) * 10

	for i := 1; i <= length; i++ {
		if i%10 == 0 {
			b1.WriteString(fmt.Sprintf("%10d", i))
			if i < length {
				b2.WriteString("\u253c")
			}
		} else if i%5 == 0 {
			b2.WriteString("\u2534")
		} else {
			b2.WriteString("\u2500")
		}
	}
	b2.WriteString("\u2524")

	fmt.Printf("%30.30s %s\n", " ", b1.String())
	fmt.Printf("%30.30s %s\n", " ", b2.String())

	for _, k := range keys {
		s := w.timeline(length, w.stats[k])
		fmt.Printf("%-30.30s %v\n", k, s)
	}
}

func (w *TheWeb) timeline(length int, v []int) string {
	t := 0
	b := strings.Builder{}
	for _, ts := range v {
		for i := t; i < ts; i++ {
			if i%10 == 0 {
				b.WriteString("\u2502")
			} else {
				b.WriteString(" ")
			}
			t++
		}
		b.WriteString("\u25cf")
		t++
	}

	for i := t; i < length; i++ {
		if i%10 == 0 {
			b.WriteString("\u2502")
		} else {
			b.WriteString(" ")
		}
	}
	b.WriteString("\u2502")

	return b.String()
}
