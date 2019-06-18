package main

import (
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"
	"strings"
)

func FindAll(executable string) ([]int, error) {
	procs, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	var matching_procs []int
	for _, proc := range procs {
		if proc.Executable() == executable {
			matching_procs = append(matching_procs, proc.Pid())
		}
	}
	return matching_procs, nil
}

func FindProcByCmd(procs []*process.Process, substr string) (*process.Process, error) {
	for _, proc := range procs {
		cmd, err := proc.Cmdline()
		if err != nil {
			return nil, err
		}
		if strings.Contains(cmd, substr) {
			return proc, nil
		}
	}
	return nil, nil
}
