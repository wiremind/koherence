package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/bs"
	"github.com/wiremind/koherence/cmd/koherence/check"
	"github.com/wiremind/koherence/machine"
)

var serveCommand = cli.Command{
	Name:   "serve",
	Usage:  "serve utilities",
	Action: serveAllCommand,
}

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "pong")
}

func debugBsHandler(w http.ResponseWriter, req *http.Request) {
	var err error

	machineInfos, err := machine.ReadFsInfos()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	infos := bs.ExtractBsInfos(machineInfos)

	b, err := json.Marshal(infos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func debugMachineFsHandler(w http.ResponseWriter, req *http.Request) {
	var infos *machine.MachineInfos
	var err error

	infos, err = machine.ReadFsInfos()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(infos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func debugMachineOpenstackHandler(w http.ResponseWriter, req *http.Request) {
	var infos *machine.MachineInfos
	var err error

	infos, err = machine.ReadOpenstackInfos()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(infos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func debugOpenstackHandler(w http.ResponseWriter, req *http.Request) {
	infos, err := machine.ReadFsInfos()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bs, err := bs.OpenstackGetBlockStorage(infos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(bs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func checkBsHandler(w http.ResponseWriter, req *http.Request) {
	var err error

	machineInfos, err := check.MachineChecker()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bsFs := bs.ExtractBsInfos(machineInfos)

	bsProvider, err := check.GetBsProvider(machineInfos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, _ := check.BsMerge(machineInfos, bsFs, bsProvider)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func checkMachineHandler(w http.ResponseWriter, req *http.Request) {
	_, err := check.MachineChecker()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func checkOpenstackMultiattachHandler(w http.ResponseWriter, req *http.Request) {
	var err error

	machineInfos, err := check.MachineChecker()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	multiAttachments, err := bs.OpenstackGetMultiAttachments(machineInfos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(multiAttachments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func serveAllCommand(clicontext *cli.Context) error {
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/debug/bs", debugBsHandler)
	http.HandleFunc("/debug/machine/fs", debugMachineFsHandler)
	http.HandleFunc("/debug/machine/openstack", debugMachineOpenstackHandler)
	http.HandleFunc("/debug/openstack", debugOpenstackHandler)
	http.HandleFunc("/check/bs", checkBsHandler)
	http.HandleFunc("/check/machine", checkMachineHandler)
	http.HandleFunc("/check/openstack/multiattach", checkOpenstackMultiattachHandler)
	slog.Info("listen port :8080")
	return http.ListenAndServe(":8080", nil)
}
