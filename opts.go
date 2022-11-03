package main

type Opt int

const (
	FailUnsetBusyHost Opt = iota
)

func (o Opt) IsIn(opts []Opt) bool {
	for _, v := range opts {
		if v == o {
			return true
		}
	}
	return false
}
