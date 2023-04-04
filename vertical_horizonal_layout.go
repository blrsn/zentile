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
	wx, wy, ww, wh := state.WorkAreaDimensions(l.WorkspaceNum)

	// concatenating all the window clients into a single list
	// The master window proportional zoom is not supported in this layout.
	// Selecting a window as a master using the hotkey has the effect of
	// 		swapping it with whatever window was currently  in the number
	//		one (top-left) position.
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
	// i have a 1080x1920 (portrait orientation) on the left
	// and a 1920x1080 (landscape) on the right.  EWMH sees one workarea of
	// rect 3000x1920, so there are non-visible spaces in the workarea where
	// windows may (but shouldn't) be placed.  These sections are to the
	// right of my portrait monitor in the regions above and below
	// my landscape monitor's configured vertical offsets
	// ( x>=1080 && (y<630 || y>1710) )
	//
	// My largest visible box that spans both screens is a rect in the EWMH
	// workarea with a top-left origin at (0,630) and dimensions 3000x1080
	// (i.e. (1080+1920)x1080 )

	gap := Config.Gap
	colsize := ww/cols - gap

	height := Config.SqHeight
	if height == 0 {
		height = wh
	}
	rowsize := height/rows - gap // see TODO above

	padx := 0
	pady := 0

	if rowsize < colsize && cols > 1 {
		padx = (colsize - rowsize) * cols / (cols - 1)
		colsize = rowsize
		for padx > colsize {
			colsize = colsize * 3 / 2
			padx = (ww - colsize*cols) / (cols - 1)
		}
	} else if rowsize > colsize && rows > 1 {
		pady = (rowsize - colsize) * rows / (rows - 1)
		rowsize = colsize
		for pady > rowsize {
			rowsize = rowsize * 3 / 2
			pady = (height - rowsize*rows) / (rows - 1) // see TODO above
		}
	}

	mx := wx
	my := Config.SqCeiling + wy // see TODO above

	log.Info("cols: ", cols, " rows: ", rows, " colsize: ", colsize, " rowsize: ", rowsize, " padx: ", padx, " pady: ", pady)

	currcol := 1
	for _, c := range allClients {
		if Config.HideDecor {
			c.UnDecorate()
		}

		log.Info("Moving ", c.name(), ": ", " X: ", mx, " Y: ", my)
		c.MoveResize(mx, my, colsize, rowsize)

		mx = mx + colsize + padx + gap
		currcol = currcol + 1
		if currcol > cols {
			mx = wx
			my = my + rowsize + pady + gap
			currcol = 1
		}
	}

	state.X.Conn().Sync()
}
