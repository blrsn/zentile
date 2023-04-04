package main

import (
	"math"

	"github.com/blrsn/zentile/state"
	log "github.com/sirupsen/logrus"
)

type VerticalLayout struct {
	*VertHorz
}

func (l *VerticalLayout) Do() {
	log.Info("Switching to Vertical Layout")
	wx, wy, ww, wh := state.WorkAreaDimensions(l.WorkspaceNum)
	msize := len(l.masters)
	ssize := len(l.slaves)

	mx := wx
	mw := int(float64(ww) * l.Proportion)
	sx := mx + mw
	sw := ww - mw
	gap := Config.Gap

	if msize > 0 {
		mh := (wh - (msize+1)*gap) / msize
		if ssize == 0 {
			mw = ww
		}

		for i, c := range l.masters {
			if Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(mx+gap, gap+wy+i*(mh+gap), mw-2*gap, mh)
		}
	}

	if ssize > 0 {
		sh := (wh - (ssize+1)*gap) / ssize
		if msize == 0 {
			sx, sw = wx, ww
		}

		for i, c := range l.slaves {
			if Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(sx, gap+wy+i*(sh+gap), sw-gap, sh)
		}
	}

	state.X.Conn().Sync()
}

type HorizontalLayout struct {
	*VertHorz
}

func (l *HorizontalLayout) Do() {
	log.Info("Switching to Horizontal Layout")
	wx, wy, ww, wh := state.WorkAreaDimensions(l.WorkspaceNum)
	msize := len(l.masters)
	ssize := len(l.slaves)

	my := wy
	mh := int(float64(wh) * l.Proportion)
	sy := my + mh
	sh := wh - mh
	gap := Config.Gap

	if msize > 0 {
		mw := (ww - (msize+1)*gap) / msize
		if ssize == 0 {
			mh = wh
		}

		for i, c := range l.masters {
			if Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(gap+wx+i*(mw+gap), my+gap, mw, mh-2*gap)
		}
	}

	if ssize > 0 {
		sw := (ww - (ssize+1)*gap) / ssize
		if msize == 0 {
			sy, sh = wy, wh
		}

		for i, c := range l.slaves {
			if Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(gap+wx+i*(sw+gap), sy, sw, sh-gap)
		}
	}

	state.X.Conn().Sync()
}

type SquareLayout struct {
	*VertHorz
}

func (l *SquareLayout) Do() {
	// intended for lots of small windows
	log.Info("Switching to Square Layout")
	wx, wy, ww, _ := state.WorkAreaDimensions(l.WorkspaceNum)

	allClients := append(l.masters, l.slaves...)
	csize := len(allClients)

	if csize == 0 {
		state.X.Conn().Sync()
		return
	}

	cols := int(math.Floor(math.Sqrt(float64(csize))))

	// cols^2 + 2*cols + 1 === (cols + 1)^2
	extraRows := int(math.Ceil(float64(csize)/float64(cols))) - cols // 0..2

	rows := cols // default to perfect square
	if extraRows > 0 {
		rows = rows + 1
	}
	if extraRows == 2 {
		cols = cols + 1
	}

	// TODO: properly support multi-monitor geometry
	colsize := ww / cols
	rowsize := 1080 / rows

	sizediff := int(math.Abs(float64(colsize-rowsize))) / 2
	if colsize < rowsize {
		colsize = colsize + sizediff
		rowsize = rowsize - sizediff
	} else if rowsize < colsize {
		rowsize = rowsize + sizediff
		colsize = colsize - sizediff
	}

	log.Info("cols: ", cols, " rows: ", rows, " colsize: ", colsize, " rowsize: ", rowsize)

	// TODO: properly support multi-monitor geometry
	mx := wx
	my := 630 + wy

	for _, c := range allClients {
		if Config.HideDecor {
			c.UnDecorate()
		}

		// nx := mx + colsize
		// if nx > ww+wx {
		// 	mx = wx
		// 	my = my + rowsize
		// }

		log.Info("Moving ", c.name(), ": ", " X: ", mx, " Y: ", my)
		c.MoveResize(mx, my, colsize, rowsize)
		mx = mx + colsize
		if mx > ww+wx {
			mx = wx
			my = my + rowsize
		}
	}

	state.X.Conn().Sync()
}
