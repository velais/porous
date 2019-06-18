package main

import (
	"fmt"
	"github.com/kevinburke/ssh_config"
	"github.com/shirou/gopsutil/process"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type AppState struct {
	tunnels []*Tunnel
}

type Tunnel struct {
	Host        string
	Hostname    string
	Forward     string
	ForwardKind string
	State       string
	Raw         []ssh_config.Node
	Proc        *process.Process
}

func NewAppState() (*AppState, error) {
	tunnels, err := load()
	if err != nil {
		return nil, err
	}

	as := AppState{
		tunnels: tunnels,
	}
	return &as, nil
}

func (as *AppState) ReloadTunnels() error {
	tunnels, err := load()
	if err != nil {
		return err
	}
	as.tunnels = tunnels
	return nil
}

func (as *AppState) GetTunnels() []*Tunnel {
	return as.tunnels
}

func load() ([]*Tunnel, error) {
	pids, err := FindAll("ssh")
	if err != nil {
		return nil, err
	}

	var procs []*process.Process
	for _, pid := range pids {
		proc, _ := process.NewProcess(int32(pid))
		procs = append(procs, proc)
	}

	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
	if err != nil {
		return nil, fmt.Errorf("could not open ssh config\n %s", err)
	}

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode ssh config:\n %s", err)
	}


	r := regexp.MustCompile(".*(LocalForward|RemoteForward).*")

	var tunnels []*Tunnel
	for _, host := range cfg.Hosts {
		var fwdNodes []*ssh_config.KV
		for _, node := range host.Nodes {
			if r.MatchString(node.String()) {
				switch t := node.(type) {
				case *ssh_config.KV:
					fwdNodes = append(fwdNodes, t)
				default:
					continue
				}
			}
		}

		if len(fwdNodes) > 0 {
			alias := host.Patterns[0].String()
			hostname, err := cfg.Get(host.Patterns[0].String(), "Hostname")

			if err != nil {
				break
			}

			var fwdStrs []string
			for _, fwd := range fwdNodes {
				fwdStrs = append(fwdStrs, fmt.Sprintf("-%s %s", string(fwd.Key[0]), fwd.Value))
			}



			proc, err := FindProcByCmd(procs, alias)
			if err != nil {
				return nil, err
			}
			tunnel := Tunnel{
				Host:     alias,
				Hostname: hostname,
				Forward: strings.Join(fwdStrs, " "),
				Raw: host.Nodes,
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
	return tunnels, nil
}
