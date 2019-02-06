package main

import (
	"flag"
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"github.com/kr/pty"
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
	if tunnel.State == "Open" {
		stateColor = "green"
	}
	return "[\u25A3](fg:" + stateColor + ") " + tunnel.Host + " " + tunnel.Forward
}

func renderTunnels(tunnels []*Tunnel) []string {
	rows := []string{}
	for _, tun := range tunnels {
		rows = append(rows, renderTunnelString(tun))
	}
	return rows
}

func render(as *AppState) {

	l := widgets.NewList()
	l.SelectedRow = 0
	l.SelectedRowStyle = ui.Style{Fg: ui.ColorYellow, Bg: l.TextStyle.Bg}
	l.SetRect(0, 0, 50, 8)

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
			ui.Render(l)

		case "j":
			l.ScrollDown()
			ui.Render(l)

		case "x":
			selected := as.GetTunnels()[l.SelectedRow]
			if selected.State == "Open" {
				err := selected.Proc.Kill()
				if err != nil {}
				as.ReloadTunnels()
				l.Rows = renderTunnels(as.GetTunnels())
				ui.Clear()
				ui.Render(l)
			}

		case "o":
			selected := as.GetTunnels()[l.SelectedRow]
			if selected.State == "Closed" {
				openTunnel(selected)
				as.ReloadTunnels()
				l.Rows = renderTunnels(as.GetTunnels())
				ui.Clear()
				ui.Render(l)
			}

		case "r":
			as.ReloadTunnels()
			l.Rows = renderTunnels(as.GetTunnels())
			ui.Render(l)
		}
	}

}

func openTunnel(tunnel *Tunnel)  {
	cmd := exec.Command("ssh", "-fN", tunnel.Host)
	ui.Clear()
	ui.Render()

	tty, err := pty.Start(cmd)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		tty.Close()
	}()

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
}



