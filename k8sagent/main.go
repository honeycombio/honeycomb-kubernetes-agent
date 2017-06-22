package main

import (
	"fmt"
	"log"

	"github.com/honeycomb/honeycomb-kubernetes-agent/k8sagent/state"
)

func main() {
	record := state.NewRecord()

	snap, err := state.NewSnapshotter(
		record,
		"/Users/alex/src/go/src/github.com/honeycomb/honeycomb-kubernetes-agent/kubeconfig",
		"ip-10-0-13-104.us-west-2.compute.internal",
		"app=nginx")
	if err != nil {
		log.Fatal(err)
	}

	if err := snap.Snapshot(); err != nil {
		log.Fatal(err)
	}

	for _, pod := range record.Pods().Items {
		fmt.Println(pod.Name)
	}
}
