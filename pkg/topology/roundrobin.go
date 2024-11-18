package topology

import (
	"errors"
)

type RoundRobinBalance struct {
	curIndex int
	rss      [][]string
}

func (r *RoundRobinBalance) Add(params ...[]string) error {
	if len(params) == 0 {
		return errors.New("params len 1 at least")
	}
	addr := params[0]
	r.rss = append(r.rss, addr)

	return nil
}

func (r *RoundRobinBalance) Next() []string {
	if len(r.rss) == 0 {
		return []string{}
	}
	lens := len(r.rss)
	if r.curIndex >= lens {
		r.curIndex = 0
	}
	curAdd := r.rss[r.curIndex]
	r.curIndex = (r.curIndex + 1) % lens
	return curAdd
}
