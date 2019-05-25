package main

import (
	"fmt"
	. "github.com/gizak/termui/v3"
	"image"
)

type Menu struct {
	Block
	Rows []*Tunnel
	SelectedRow uint
	rowStyle Style
}

func NewMenu() *Menu {
	return &Menu{
		Block: *NewBlock(),
		rowStyle: NewStyle(ColorWhite),
	}
}

func (self *Menu) TunnelToRow (tunnel *Tunnel) []Cell {
	stateColor := "red"
	stateShape := "\u25A3"
	if tunnel.State == "Open" {
		stateColor = "green"
		stateShape = "\u25C8"
	}

	rowString := fmt.Sprintf("[%s](fg:%s) %-30v %-40v %-40v",
		stateShape, stateColor, tunnel.Host, tunnel.Forward, tunnel.Hostname)
	cells := ParseStyles(rowString, self.rowStyle)

	if len(cells) < self.Inner.Max.X {
		padding := make([]Cell, self.Inner.Max.X - len(cells))
		for i := range padding {
			padding[i] = NewCell(0)
		}
		cells = append(cells, padding...)
	}

	return cells
}

func (self *Menu) Draw(buf * Buffer) {
	self.Block.Draw(buf)

	point := self.Inner.Min

	for row := uint(0); row < uint(len(self.Rows)) && point.Y < self.Inner.Max.Y; row ++ {
		tunnel := self.Rows[row]
		cells := self.TunnelToRow(tunnel)

		for j := 0; j < len(cells) && point.Y < self.Inner.Max.Y; j++ {
			style := cells[j].Style
			if row == self.SelectedRow && j >= 2 {
				style = Style{Bg: ColorWhite}
			}

			buf.SetCell(NewCell(cells[j].Rune, style), point)
			point = point.Add(image.Pt(1, 0))

		}
		point = image.Pt(self.Inner.Min.X, point.Y+1)
	}
}

func (self *Menu) ScrollBy(amount int) {
	if len(self.Rows)-int(self.SelectedRow) <= amount {
		self.SelectedRow = uint(len(self.Rows) - 1)
	} else if int(self.SelectedRow)+amount < 0 {
		self.SelectedRow = 0
	} else {
		self.SelectedRow += uint(amount)
	}
}

func (self *Menu) ScrollUp() {
	self.ScrollBy(-1)
}

func (self *Menu) ScrollDown() {
	self.ScrollBy(1)
}

