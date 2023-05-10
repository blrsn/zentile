package main

import (
	"github.com/blrsn/zentile/state"
)

type Store struct {
	allowedMasters  int
	masters, slaves []Client
}

func buildStore() *Store {
	return &Store{allowedMasters: 1,
		masters: make([]Client, 0),
		slaves:  make([]Client, 0),
	}
}

func (st *Store) Add(c Client) {
	if len(st.masters) < st.allowedMasters {
		st.masters = append(st.masters, c)
	} else {
		st.slaves = append(st.slaves, c)
	}
}

func (st *Store) Remove(c Client) {
	for i, m := range st.masters {
		if m.window.Id == c.window.Id {
			if len(st.slaves) > 0 {
				st.masters[i] = st.slaves[0]
				st.slaves = st.slaves[1:]
			} else {
				st.masters = removeElement(st.masters, i)
			}
			return
		}
	}

	for i, s := range st.slaves {
		if s.window.Id == c.window.Id {
			st.slaves = removeElement(st.slaves, i)
			return
		}
	}
}

func removeElement(s []Client, i int) []Client {
	return append(s[:i], s[i+1:]...)
}

func (st *Store) IncMaster() {
	if len(st.slaves) > 1 {
		st.allowedMasters = st.allowedMasters + 1
		st.masters = append(st.masters, st.slaves[0])
		st.slaves = st.slaves[1:]
	}
}

func (st *Store) DecreaseMaster() {
	if len(st.masters) > 1 {
		st.allowedMasters = st.allowedMasters - 1
		mlen := len(st.masters)
		st.slaves = append([]Client{st.masters[mlen-1]}, st.slaves...)
		st.masters = st.masters[:mlen-1]
	}
}

func (st *Store) MakeMaster(c Client) {
	for i, slave := range st.slaves {
		if slave.window.Id == c.window.Id {
			st.masters[0], st.slaves[i] = st.slaves[i], st.masters[0]
		}
	}
}

func (st *Store) DemoteWindow(c Client) {
	// this is intended to be the functional inverse of MakeMaster above
	// instead of swapping to the front of the list, we are swapping to the end
	// one diff is we are iterating over all windows not just the masters slice
	slen := len(st.slaves)
	if slen > 0 && st.slaves[slen-1].window.Id == c.window.Id {
		// Demoting the last client promotes it
		s := st.slaves[slen-1]
		mlen := len(st.masters)
		if mlen == 0 {
			st.masters = []Client{s}
			st.slaves = st.slaves[:slen-1]
		} else {
			st.masters[0], st.slaves[slen-1] = st.slaves[slen-1], st.masters[0]
		}
		return
	}

	for i, master := range st.masters {
		if master.window.Id == c.window.Id {
			if slen > 0 {
				st.slaves[slen-1], st.masters[i] = st.masters[i], st.slaves[slen-1]
			} else {
				st.slaves = []Client{st.masters[i]}
				st.masters = append(st.masters[:i], st.masters[i+1:]...)
			}
			return
		}
	}

	if slen < 2 {
		return
	}

	for i, slave := range st.slaves {
		if slave.window.Id == c.window.Id {
			st.slaves[i], st.slaves[slen-1] = st.slaves[slen-1], st.slaves[i]
			return
		}
	}
}

func (st *Store) All() []Client {
	return append(st.masters, st.slaves...)
}

func (st *Store) Next() Client {
	clients := st.All()
	lastIndex := len(clients) - 1

	for i, c := range clients {
		if c.window.Id == state.ActiveWin {
			next := i + 1
			if next > lastIndex {
				next = 0
			}
			return clients[next]
		}
	}

	return Client{}
}

func (st *Store) Previous() Client {
	clients := st.All()
	lastIndex := len(clients) - 1

	for i, c := range clients {
		if c.window.Id == state.ActiveWin {
			prev := i - 1
			if prev < 0 {
				prev = lastIndex
			}
			return clients[prev]
		}
	}

	return Client{}
}
