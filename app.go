package main

import (
	"github.com/kevinburke/ssh_config"
	"github.com/shirou/gopsutil/process"
	"os"
	"path/filepath"
	"regexp"

)

type AppState struct {
	tunnels []*Tunnel
}

type Tunnel struct {
	Host string
	Hostname string
	Forward string
	State string
	Proc  *process.Process

}

func NewAppState() *AppState  {
	as := AppState{
		tunnels: load(),
	}
	return &as
}

func (as *AppState) ReloadTunnels() {
	as.tunnels = load()
}

func (as *AppState) GetTunnels() []*Tunnel {
	return as.tunnels
}

func load() []*Tunnel {
	pids := FindAll("ssh")
	var procs []*process.Process
	for _, pid := range pids {
		proc, _ := process.NewProcess(int32(pid))
		procs = append(procs, proc)
	}
	f, _ := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
	cfg, _ := ssh_config.Decode(f)

	r := regexp.MustCompile(".*(LocalForward|RemoteForward).*")

	var tunnels []*Tunnel
	for _, host := range cfg.Hosts {
		for _, node := range host.Nodes {
			if r.MatchString(node.String()) {

				alias := host.Patterns[0].String()
				hostname, err := cfg.Get(host.Patterns[0].String(), "Hostname")
				if err != nil {
					break
				}

				localFwd, err := cfg.Get(alias, "LocalForward")
				remoteFwd, err := cfg.Get(alias, "RemoteForward")


				proc := FindProcs(procs, alias)
				tunnel := Tunnel{
					Host: alias,
					Hostname: hostname,
					Forward: localFwd + remoteFwd,
				}

				if proc != nil {
					tunnel.Proc = proc
					tunnel.State = "Open"
				} else {
					tunnel.State = "Closed"
				}


				tunnels = append(tunnels, &tunnel)
			}
		}
	}

	return tunnels
}




















