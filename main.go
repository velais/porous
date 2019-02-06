package main

import (
	"flag"
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"github.com/kr/pty"
	"github.com/nsf/termbox-go"
	"io"
	"log"
	"os"
	"os/exec"
)

var (
	version = "0.0.1"
	versionStr = fmt.Sprintf("porous %v", version)
)

func main() {

	var (
		versionFlag = flag.Bool("v", false, "output version information")
		helpFlag    = flag.Bool("h", false, "display this help message")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Println(versionStr)
		os.Exit(0)
	}

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	//UI
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	appState := NewAppState()

	render(appState)
}

var helpMsg = `porous - ssh tunnel manager

usage: porous [options]

options:`

func printHelp() {
	fmt.Println(helpMsg)
	flag.PrintDefaults()
}

func renderTunnelString(tunnel *Tunnel) string {
	stateColor := "red"
	stateShape := "\u25A3"
	if tunnel.State == "Open" {
		stateColor = "green"
		stateShape = "\u25C8"
	}
	return "[" + stateShape + "](fg:" + stateColor + ") " + tunnel.Host + " " + tunnel.Forward
}

func renderTunnels(tunnels []*Tunnel) []string {
	rows := []string{}
	for _, tun := range tunnels {
		rows = append(rows, renderTunnelString(tun))
	}
	return rows
}

func reloadSelectedStyle(as *AppState, l *widgets.List) {
	selected := as.GetTunnels()[l.SelectedRow]
	if selected.State == "Open" {
		l.SelectedRowStyle = ui.Style{Bg: ui.ColorGreen}
	} else {
		l.SelectedRowStyle = ui.Style{Bg: ui.ColorRed}
	}
}

func render(as *AppState) {
	width, height := ui.TerminalDimensions()

	l := widgets.NewList()
	l.Border = false
	l.SelectedRow = 0
	l.SelectedRowStyle = ui.Style{Bg: ui.ColorWhite, Modifier: ui.ModifierClear}
	l.SetRect(0, -1, width, height)

	l.Rows = renderTunnels(as.GetTunnels())
	ui.Render(l)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {

		case "q":
			return

		case "k":
			l.ScrollUp()
			reloadSelectedStyle(as, l)
			ui.Render(l)

		case "j":
			l.ScrollDown()
			reloadSelectedStyle(as, l)
			ui.Render(l)

		case "x":
			selected := as.GetTunnels()[l.SelectedRow]
			if selected.State == "Open" {
				err := selected.Proc.Kill()
				if err != nil {}
				as.ReloadTunnels()
				reloadSelectedStyle(as, l)
				l.Rows = renderTunnels(as.GetTunnels())
				ui.Clear()
				ui.Render(l)
			}

		case "o":
			selected := as.GetTunnels()[l.SelectedRow]
			if selected.State == "Closed" {
				openTunnel(selected)
				as.ReloadTunnels()
				reloadSelectedStyle(as, l)
				l.Rows = renderTunnels(as.GetTunnels())
				ui.Clear()
				ui.Render(l)
			}

		case "r":
			as.ReloadTunnels()
			reloadSelectedStyle(as, l)
			l.Rows = renderTunnels(as.GetTunnels())
			ui.Render(l)
		}
	}

}

func openTunnel(tunnel *Tunnel)  {
	cmd := exec.Command("ssh", "-fN", tunnel.Host)
	ui.Clear()
	ui.Render()

	_= termbox.Sync()

	tty, err := pty.Start(cmd)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() { _ = tty.Close() }()

	go func() {
		io.Copy(os.Stdout, tty)
		io.Copy(os.Stderr, tty)
	}()
	go func() {
		io.Copy(tty, os.Stdin)
	}()

	err = cmd.Wait()
	if err != nil {
		log.Fatalln(err)
	}

	_= termbox.Sync()
}



