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

	gap := Config.Gap

	// concatenating all the window clients into a single list
	// The master window proportional zoom is not supported in this layout.
	// Selecting a window as a master using the hotkey has the effect of
	// 		swapping it with whatever window was currently  in the number
	//		one (top-left) position.
	allClients := append(l.masters, l.slaves...)

	// sub-regions of the main work area (to describe multiple monitors)
	// regions is a slice of (x,y,width,height) arrays
	// clients get divided evenly among regions then regions are
	// rendered serially
	var regions [][4]int = Config.MMRegions

	if len(regions) == 0 {
		regions = append(regions, [4]int{wx, wy, ww, wh})
	}

	nregions := len(regions)
	segsize := len(allClients) / nregions

	for i := 0; i < nregions; i += 1 {
		region := regions[i]
		rx := region[0]
		ry := region[1]
		rw := region[2]
		rh := region[3]

		var regionClients []Client
		if i+1 == nregions {
			regionClients = allClients[i*segsize:]
		} else {
			regionClients = allClients[i*segsize : (i+1)*segsize]
		}
		csize := len(regionClients)

		if csize == 0 {
			continue
		}
		cols := int(math.Floor(math.Sqrt(float64(csize))))
		rows := cols // default to perfect square

		// cols^2 + 2*cols + 1 === (cols + 1)^2
		extras := int(math.Ceil(float64(csize)/float64(cols))) - cols // 0..2

		// if taller than wide, add additional rows first
		// if wider than tall, add additional columns first
		if extras > 0 {
			if rh >= rw {
				rows = rows + 1
			} else {
				cols = cols + 1
			}
		}
		if extras == 2 {
			if rh >= rw {
				cols = cols + 1
			} else {
				rows = rows + 1
			}
		}

		colsize := rw/cols - gap
		rowsize := rh/rows - gap

		padx := 0
		pady := 0

		// here is an algo for auto-padding (gap becomes minimum pad)
		// i ended up not liking it and figured it would add too
		// much complexity to the config, but here it is...

		// if rowsize < colsize && cols > 1 {
		// 	padx = (colsize - rowsize) * cols / (cols - 1)
		// 	colsize = rowsize
		// 	for padx > colsize {
		// 		colsize = colsize * 3 / 2
		// 		padx = (rw - colsize*cols) / (cols - 1)
		// 	}
		// } else if rowsize > colsize && rows > 1 {
		// 	pady = (rowsize - colsize) * rows / (rows - 1)
		// 	rowsize = colsize
		// 	for pady > rowsize {
		// 		rowsize = rowsize * 3 / 2
		// 		pady = (rh - rowsize*rows) / (rows - 1)
		// 	}
		// }

		mx := rx
		my := ry

		log.Info("cols: ", cols, " rows: ", rows, " colsize: ", colsize, " rowsize: ", rowsize, " padx: ", padx, " pady: ", pady)

		currcol := 1
		for _, c := range regionClients {
			if Config.HideDecor {
				c.UnDecorate()
			}

			log.Info("Moving ", c.name(), ": ", " X: ", mx, " Y: ", my)
			c.MoveResize(mx, my, colsize, rowsize)

			mx = mx + colsize + padx + gap
			currcol = currcol + 1
			if currcol > cols {
				mx = rx
				my = my + rowsize + pady + gap
				currcol = 1
			}
		}
	}

	state.X.Conn().Sync()
}
