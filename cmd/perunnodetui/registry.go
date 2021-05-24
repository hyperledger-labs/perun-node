package main

import "sync"

// registry stores the mapping of channels to rows in the table.
// channels can by looked up by both row number as well as channel ID.
type registry struct {
	sync.Mutex
	proposals []*channel
	byChID    map[string]int
}

func newRegistry(size int) *registry {
	return &registry{
		proposals: make([]*channel, size),
		byChID:    make(map[string]int),
	}
}

func (r *registry) putAtIndex(index int, p *channel) {
	r.Lock()
	if index < len(r.proposals) {
		r.proposals[index] = p
	}
	r.Unlock()
}

func (r *registry) getByIndex(index int) (p *channel) {
	r.Lock()
	defer r.Unlock()
	if index >= len(r.proposals) {
		return nil
	}
	return r.proposals[index]
}

func (r *registry) setChID(chID string, sNo int) {
	r.Lock()
	r.byChID[chID] = sNo
	r.Unlock()
}

func (r *registry) getByChID(chID string) (p *channel) {
	r.Lock()
	defer r.Unlock()
	index, ok := r.byChID[chID]
	if !ok {
		return nil
	}
	return r.proposals[index]
}
