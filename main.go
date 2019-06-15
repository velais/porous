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
	"log"
	"os"
	"os/exec"
)

var (
	version    = "0.0.1"
	versionStr = fmt.Sprintf("porous %v", version)

	stderrLogger = log.New(os.Stderr, "", 0)
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
		stderrLogger.Fatalf("failed to initialize porous: %v", err)
	}

	appState, err := NewAppState()
	if err != nil {
		stderrLogger.Fatalf("failed to initialize porous: %v", err)
	}

	//UI
	err = ui.Init()
	if err != nil {
		stderrLogger.Fatalf("failed to initialize termui: %v", err)
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

func errExit(msg string, err error) {
	renderClear()
	stderrLogger.Fatalf("%s: %v", msg, err)
}

func renderClear() {
	ui.Clear()
	ui.Render()
	_ = termbox.Sync()
}

func renderMenu(as *AppState) {
	width, height := ui.TerminalDimensions()

	menu := NewMenu()
	menu.SelectedRow = 0
	menu.SetRect(-1, -1, width, height) //Place at -1 to offset internal padding
	menu.Border = false
	menu.Rows = as.GetTunnels()

	ui.Render(menu)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q":
			return
		case "k", "<Up>":
			menu.ScrollUp()
			ui.Render(menu)
		case "j", "<Down>":
			menu.ScrollDown()
			ui.Render(menu)
		case "x":
			selected := as.GetTunnels()[menu.SelectedRow]
			if selected.State == "Open" {
				_ = selected.Proc.Kill()
				reloadMenu(as, menu)
			}
		case "o", "<Enter>":
			selected := as.GetTunnels()[menu.SelectedRow]
			if selected.State == "Closed" {
				renderClear()
				if err := openTunnel(selected); err != nil {
					renderOpenError(uiEvents, err)
				}
				reloadMenu(as, menu)
			}
		case "r":
			reloadMenu(as, menu)
		case "i":
			selected := as.GetTunnels()[menu.SelectedRow]
			renderInfo(uiEvents, selected)
			reloadMenu(as, menu)
		}

		if e.Type == ui.ResizeEvent {
			width, height = ui.TerminalDimensions()
			menu.SetRect(0, -1, width, height)
			menu.Rows = as.GetTunnels()
			ui.Render(menu)
		}
	}

}

func reloadMenu(as *AppState, menu *Menu) {
	if err := as.ReloadTunnels(); err != nil {
		errExit("failed to reload", err)
	}
	menu.Rows = as.GetTunnels()
	ui.Clear()
	ui.Render(menu)
}

func renderInfo(uiEvents <-chan ui.Event, tunnel *Tunnel) {
	width, height := ui.TerminalDimensions()

	t := "Host " + tunnel.Host + "\n"
	for _, node := range tunnel.Raw {
		t = t + node.String() + "\n"
	}

	p := widgets.NewParagraph()
	p.SetRect(-1, -1, width, height) //Place at -1 to offset internal padding
	p.Border =  false
	p.Text = t

	ui.Clear()
	ui.Render(p)

	for {
		e := <-uiEvents
		switch e.ID {
		case "q":
			return
		}
	}
}

func renderOpenError(uiEvents <-chan ui.Event, err error) {
	width, height := ui.TerminalDimensions()

	p := widgets.NewParagraph()
	p.Text = fmt.Sprintf("Tunnel failed to open with:\n" +
		"[%s](fg:red)\n" +
		"Press <Enter> to continue", err.Error())
	p.SetRect(-1, -1, width, height) //Place at -1 to offset internal padding
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
