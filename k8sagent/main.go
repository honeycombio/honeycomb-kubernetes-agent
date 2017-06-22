package main

import (
	"fmt"
	"log"
	"os"

	"github.com/honeycomb/honeycomb-kubernetes-agent/k8sagent/api"
	"github.com/honeycomb/honeycomb-kubernetes-agent/k8sagent/state"
)

func main() {
	// Get name of node this daemon is running on. This is usually
	// passed in via `fieldPath` from Kubernetes.
	nodeName := os.Getenv("NODE_NAME")

	record := state.NewRecord()

	// TODO: Find this in ConfigMap or command args.
	config, err := api.ReadFromFile("k8sagent/testdata/test.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Find local name programmatically, get file from ConfigMap
	// or command args.
	snap, err := state.NewSnapshotter(
		config,
		record,
		"/Users/alex/src/go/src/github.com/honeycomb/honeycomb-kubernetes-agent/kubeconfig",
		nodeName)
	if err != nil {
		log.Fatal(err)
	}

	if err := snap.Snapshot(); err != nil {
		log.Fatal(err)
	}

	datasetName := "kubernetestest"
	pods, ok := record.Pods(datasetName)
	if !ok {
		log.Fatalf("Could not find entry associated with dataset '%s'", datasetName)
	}

	for _, pod := range pods.Items {
		fmt.Printf("%s %s\n", pod.Name, pod.UID)
	}
}
