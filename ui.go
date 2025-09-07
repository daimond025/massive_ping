package main

import (
	"fmt"
	log2 "log"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type userInterface struct {
	app   *tview.Application
	grid  *tview.Grid
	table *tview.Table
}

type cell_info struct {
	title   string
	align   int
	initVal func(destination) string
	content func(history) string
}

var coldef = [...]struct {
	title   string
	align   int
	initVal func(destination) string
	content func(*history) string
}{
	{
		title:   "host",
		align:   tview.AlignLeft,
		initVal: func(d destination) string { return d.host },
	},
	{
		title:   "address",
		align:   tview.AlignLeft,
		initVal: func(d destination) string { return d.remote.IP.String() },
	},
	{
		title:   "sent",
		align:   tview.AlignRight,
		content: func(st *history) string { return strconv.Itoa(st.send) },
	},
	{
		title:   "loss",
		align:   tview.AlignRight,
		content: func(st *history) string { return st.getLost() },
	},
	{
		title:   "last",
		align:   tview.AlignRight,
		content: func(st *history) string { return st.getLast() },
	},
	{
		title:   "best",
		align:   tview.AlignRight,
		content: func(st *history) string { return st.getBest() },
	},
	{
		title:   "worst",
		align:   tview.AlignRight,
		content: func(st *history) string { return st.getWorst() },
	},
	{
		title:   "mean",
		align:   tview.AlignRight,
		content: func(st *history) string { return st.getMean() },
	},
}
var (
	pinger_ui *Pinger
	ui_ui     *userInterface
)

func Draw(pinger *Pinger, ui *userInterface) {
	row := 1

	for remote, dst := range pinger.target {
		stats, ok_exist := pinger.history[remote]

		for col, def := range coldef {
			var cell *tview.TableCell

			if def.initVal != nil {
				cell = tview.NewTableCell(def.initVal(dst))
			} else if ok_exist && def.content != nil {
				cell = tview.NewTableCell(def.content(&stats))
			} else {
				cell = tview.NewTableCell("n/a")
			}
			ui.table.SetCell(row, col, cell.SetAlign(def.align))
		}
		row += 1
	}
}
func buildTUI(pinger *Pinger) *userInterface {
	ui := &userInterface{
		app:   tview.NewApplication(),
		table: tview.NewTable().SetBorders(true).SetFixed(2, 0),
		grid:  tview.NewGrid().SetRows(3, 0, 10).SetColumns(0),
	}

	title := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]multiping[white] press q to exit")

	logs := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	log2.SetFlags(log2.Ltime | log2.LUTC)
	log2.SetOutput(logs)

	ui.grid.AddItem(title, 0, 0, 1, 1, 0, 0, false)
	ui.grid.AddItem(ui.table, 1, 0, 1, 1, 0, 0, true)
	ui.grid.AddItem(logs, 2, 0, 1, 1, 0, 0, false)

	// setup controls
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			ui.app.Stop()
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				ui.app.Stop()
				return nil
			}
		}
		return event
	})

	//  header
	for col, def := range coldef {
		cell := tview.NewTableCell(def.title).SetAlign(def.align)
		if col == 2 {
			cell.SetExpansion(1)
		}
		ui.table.SetCell(0, col, cell)
	}
	Draw(pinger, ui)

	pinger_ui = pinger
	ui_ui = ui

	return ui
}

func (ui *userInterface) Run() error {
	ui.app.SetRoot(ui.grid, true).SetFocus(ui.table)
	return ui.app.Run()
}

func (ui *userInterface) update(pinger *Pinger, interval time.Duration) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			ui.app.QueueUpdateDraw(func() {
				Draw(pinger, ui)
			})
		}
	}
}

const tsDividend = float64(time.Millisecond) / float64(time.Nanosecond)

func ts(dur time.Duration) string {
	if 10*time.Microsecond < dur && dur < time.Second {
		return fmt.Sprintf("%0.2f", float64(dur.Nanoseconds())/tsDividend)
	}
	return dur.String()
}
