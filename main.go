package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/nsf/termbox-go"
	"io"
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

	renderMenu(appState)
}

var helpMsg = `porous - ssh tunnel manager

usage: porous [options]

options:`

func printHelp() {
	fmt.Println(helpMsg)
	flag.PrintDefaults()
}

func renderClear() {
	ui.Clear()
	ui.Render()
	_ = termbox.Sync()
}

func renderMenu(as *AppState) {
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
		case "o", "<Enter>":
			selected := as.GetTunnels()[l.SelectedRow]
			if selected.State == "Closed" {
				renderClear()
				err := openTunnel(selected)
				renderClear()

				if err != nil {
					renderError(uiEvents, err)
				}

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

func renderError(uiEvents <-chan ui.Event, err error) {
	width, height := ui.TerminalDimensions()

	p := widgets.NewParagraph()
	p.Text = fmt.Sprintf("Tunnel failed to open with:\n" +
		"[%s](fg:red)\n" +
		"Press <Enter> to continue", err.Error())
	p.SetRect(0, 0, width, height)
	p.Border = false

	ui.Clear()
	ui.Render(p)

	for {
		e := <-uiEvents
		switch e.ID {
		case "<Enter>", "q":
			return
		}
	}
}

func openTunnel(tunnel *Tunnel) error {
	var stderrBuf bytes.Buffer
	cmd := exec.Command("ssh", "-fN", tunnel.Host)
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()

	if err != nil {
		panic(err)
	}

	go func() {
		_, _ = io.Copy(&stderrBuf, stderrIn)
	}()

	err = cmd.Wait()

	if err != nil {
		return errors.New(string(stderrBuf.Bytes()))
	}

	return nil
}
