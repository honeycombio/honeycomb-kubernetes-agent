package main

import (
	"log"

	"github.com/honeycomb/honeycomb-kubernetes-agent/k8sagent/state"
)

func main() {
	snap, err := state.NewSnapshotter(
		"/Users/alex/src/go/src/github.com/honeycomb/honeycomb-kubernetes-agent/kubeconfig",
		"ip-10-0-13-104.us-west-2.compute.internal",
		"app=nginx")
	if err != nil {
		log.Fatal(err)
	}

	if err := snap.Snapshot(); err != nil {
		log.Fatal(err)
	}
}
