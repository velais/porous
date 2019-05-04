package main

import (
	"flag"
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/nsf/termbox-go"
	"os"
	"os/exec"
)

var (
	version    = "0.0.1"
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

	_, err := exec.LookPath("ssh")
	if err != nil {
		panic(err)
	}

	appState := NewAppState()

	//UI
	err = ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	render(appState)
}

var helpMsg = `porous - ssh tunnel manager

usage: porous [options]

options:`

func printHelp() {
	fmt.Println(helpMsg)
	flag.PrintDefaults()
}

func render(as *AppState) {
	width, height := ui.TerminalDimensions()

	l := NewMenu()
	l.SelectedRow = 0
	l.SetRect(0, -1, width, height)
	l.Border = false
	l.Rows = as.GetTunnels()

	ui.Render(l)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q":
			return
		case "k", "<Up>":
			l.ScrollUp()
			ui.Render(l)
		case "j", "<Down>":
			l.ScrollDown()
			ui.Render(l)
		case "x":
			selected := as.GetTunnels()[l.SelectedRow]
			if selected.State == "Open" {
				err := selected.Proc.Kill()
				if err != nil {
				}
				as.ReloadTunnels()
				l.Rows = as.GetTunnels()
				ui.Clear()
				ui.Render(l)
			}
		case "o":
			selected := as.GetTunnels()[l.SelectedRow]
			if selected.State == "Closed" {
				openTunnel(selected)
				as.ReloadTunnels()
				l.Rows = as.GetTunnels()
				ui.Clear()
				ui.Render(l)
			}
		case "r":
			as.ReloadTunnels()
			l.Rows = as.GetTunnels()
			ui.Render(l)
		}

		if e.Type == ui.ResizeEvent {
			width, height = ui.TerminalDimensions()
			l.SetRect(0, -1, width, height)
			l.Rows = as.GetTunnels()
			ui.Render(l)
		}
	}

}

func openTunnel(tunnel *Tunnel) {
	cmd := exec.Command("ssh", "-fN", tunnel.Host)
	ui.Clear()
	ui.Render()
	_ = termbox.Sync()

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	defer func() {
		ui.Clear()
		ui.Render()
		_ = termbox.Sync()
	}()

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
