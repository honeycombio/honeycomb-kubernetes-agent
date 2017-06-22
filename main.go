package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/k8sagent/state"
	"github.com/honeycombio/honeycomb-kubernetes-agent/tailer"
	libhoney "github.com/honeycombio/libhoney-go"

	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func main() {
	config, err := config.ReadFromFile("/etc/honeytail/config.yaml")
	if err != nil {
		fmt.Printf("Error reading configuration:\n\t%v\n", err)
		os.Exit(1)
	}

	err = libhoney.Init(libhoney.Config{
		WriteKey: config.WriteKey,
	})
	if err != nil {
		fmt.Printf("Error initializing Honeycomb transmission:\n\t%v\n", err)
		os.Exit(1)
	}

	if len(config.Parsers) == 0 {
		fmt.Printf("No parsers defined in the configuration!")
		os.Exit(1)
	}

	metadataProcessor := &KubernetesMetadataProcessor{}

	podRecord := getPodRecord(config)

	for index, parserConfig := range config.Parsers {
		pods, _ := podRecord.Pods(index)
		paths := getPodPaths(pods)
		fmt.Println("PATHS", index, paths)
		for _, path := range paths {
			handler := &EventsHandler{
				config: parserConfig,
				parser: &NoOpParser{},
			}
			handler.AddPostProcessor(metadataProcessor)
			handler.Init()

			// TODO should clean up channels as files go away
			out := make(chan string, 1000)

			tailer.TailFile(path, out)
			go handler.Handle(out)
		}
	}
	fmt.Println("running")
	// Hang out forever
	select {}
}

func getPodRecord(config *config.Config) state.Record {
	// Get name of node this daemon is running on. This is usually
	// passed in via `fieldPath` from Kubernetes.
	nodeName := os.Getenv("NODE_NAME")

	record := state.NewRecord(len(config.Parsers))

	// TODO: Find local name programmatically, get file from ConfigMap
	// or command args.
	snap, err := state.NewSnapshotter(config, record, nodeName)
	if err != nil {
		log.Fatal(err)
	}

	if err := snap.Snapshot(); err != nil {
		log.Fatal(err)
	}

	return record
}

func getPodPaths(pods *apiv1.PodList) []string {
	paths := make([]string, 0)
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			podUID := string(pod.UID)
			containerName := container.Name
			// TODO should be smarter about this
			path := fmt.Sprintf("/var/log/pods/%s/%s_0.log", podUID, containerName)
			paths = append(paths, path)
		}
	}
	return paths
}

type EventsHandler struct {
	config         *config.ParserConfig
	parser         Parser
	postprocessors []Processor
	builder        *libhoney.Builder
}

func (e *EventsHandler) Init() {
	e.builder = libhoney.NewBuilder()
	e.builder.Dataset = e.config.Dataset
	e.parser.Init()
	for _, postprocessor := range e.postprocessors {
		postprocessor.Init()
	}
}

func (e *EventsHandler) Handle(lines chan string) {
	// no multiline parsing yet
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return
			}
			parsed, err := e.parser.Parse(line)
			logrus.Debugln("parsed line", parsed, err)
			if err != nil {
				continue
			}
			for _, p := range e.postprocessors {
				p.Process(parsed)
			}
			fmt.Println("Sending line", parsed)
			e.builder.SendNow(parsed)
		}
	}
}

func (e *EventsHandler) AddPostProcessor(p Processor) {
	e.postprocessors = append(e.postprocessors, p)
}

type Parser interface {
	Init() error
	Parse(string) (map[string]interface{}, error)
}

type Processor interface {
	Init() error
	Process(data map[string]interface{})
}

type KubernetesMetadataProcessor struct {
	namespace string
}

func (k *KubernetesMetadataProcessor) Init() error {
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	return err
	//}

	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	return err
	//}

	//go func() {
	//	for {
	//		pods, err := clientset.CoreV1().Pods(k.namespace).List(metav1.ListOptions{})
	//		if err != nil {
	//			return
	//		}
	//		fmt.Println("Got pod metadata", pods)
	//		time.Sleep(1)
	//	}
	//}()
	return nil
}

type PodMetadata struct {
	PodID   string
	PodName string
	Labels  map[string]string
}

func (k *KubernetesMetadataProcessor) parsePodData(pod apiv1.Pod) *PodMetadata {
	return &PodMetadata{
		PodID:   string(pod.UID),
		PodName: pod.Name,
		Labels:  pod.Labels,
	}
}

func (k *KubernetesMetadataProcessor) Process(data map[string]interface{}) {
	// TODO actually map log line->pod metadata here, will need to plumb the
	// pod UID through to here, keep a map podUID->metadata
}

// Just parses the log line as JSON
type NoOpParser struct{}

func (n *NoOpParser) Init() error { return nil }
func (n *NoOpParser) Parse(line string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(line), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
