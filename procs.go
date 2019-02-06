package main

import (
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"
	"strings"
)

func FindAll(ex string) []int {

	procs, err := ps.Processes()
	if err != nil {}

	var matching_procs []int
	for _, proc := range procs {
		if proc.Executable() == ex {
			matching_procs = append(matching_procs, proc.Pid())
		}
	}

	return matching_procs
}

func FindProcs(procs []*process.Process, alias string) *process.Process {
	for _, proc := range procs {
		cmd, err := proc.Cmdline()
		if err != nil {
		}
		if strings.Contains(cmd, alias) {
			return proc
		}
	}
	return nil
}



