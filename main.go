package main

import (
	"encoding/json"
	"fmt"
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

	podRecord, err := getPodRecord(config)

	if err != nil {
		fmt.Printf("Error fetching pod list:\n\t%v\n", err)
		os.Exit(1)
	}

	for _, parserConfig := range config.Parsers {
		pods, _ := podRecord.Pods(parserConfig.LabelSelector)
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				path := getPodPath(pod, container)
				handler := &JSONLogHandler{
					config: parserConfig,
					parser: &NoOpParser{},
				}

				metadataProcessor := &KubernetesMetadataProcessor{
					podRecord: podRecord,
					pod:       pod,
					container: container,
				}
				handler.AddPostProcessor(metadataProcessor)
				handler.Init()

				tailer := tailer.NewTailer(path, handler)
				go tailer.Run()
			}
		}
	}
	fmt.Println("running")
	// Hang out forever
	select {}
}

func getPodRecord(config *config.Config) (state.Record, error) {
	// Get name of node this daemon is running on. This is usually
	// passed in via `fieldPath` from Kubernetes.
	nodeName := os.Getenv("NODE_NAME")

	record := state.NewRecord()

	// TODO: Find local name programmatically, get file from ConfigMap
	// or command args.
	selectors := make([]string, len(config.Parsers))
	for i, p := range config.Parsers {
		selectors[i] = p.LabelSelector
	}
	snap, err := state.NewSnapshotter(selectors, record, nodeName)
	if err != nil {
		return nil, err
	}

	if err := snap.Snapshot(); err != nil {
		return nil, err
	}

	return record, nil
}

func getPodPath(pod apiv1.Pod, container apiv1.Container) string {
	return fmt.Sprintf("/var/log/pods/%s/%s_0.log", string(pod.UID), container.Name)
}

type JSONLogHandler struct {
	config         *config.ParserConfig
	parser         Parser
	postprocessors []Processor
	builder        *libhoney.Builder
}

func (h *JSONLogHandler) Init() {
	h.builder = libhoney.NewBuilder()
	h.builder.Dataset = h.config.Dataset
	h.parser.Init()
	for _, postprocessor := range h.postprocessors {
		postprocessor.Init()
	}
}

type jsonLogLine struct {
	Log    string
	Stream string
	Time   string
}

func (h *JSONLogHandler) Handle(rawLines <-chan string) {
	// no multiline parsing yet
	for rawLine := range rawLines {
		line := &jsonLogLine{}
		err := json.Unmarshal([]byte(rawLine), line)
		if err != nil {
			logrus.WithError(err).Warnln("Error parsing JSON line")
			continue
		}

		parsed, err := h.parser.Parse(line.Log)
		logrus.Debugln("parsed line", parsed, err)
		if err != nil {
			continue
		}
		for _, p := range h.postprocessors {
			p.Process(parsed)
		}
		fmt.Println("Sending line", parsed)
		h.builder.SendNow(parsed)
	}
}

func (h *JSONLogHandler) AddPostProcessor(p Processor) {
	h.postprocessors = append(h.postprocessors, p)
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
	podRecord state.Record
	pod       apiv1.Pod
	container apiv1.Container
}

func (k *KubernetesMetadataProcessor) Init() error {
	return nil
}

func (k *KubernetesMetadataProcessor) Process(data map[string]interface{}) {
	data["kubernetes.pod"] = k.pod
	data["kubernetes.container"] = k.container
}

// Doesn't do any parsing
type NoOpParser struct{}

func (p *NoOpParser) Init() error { return nil }
func (p *NoOpParser) Parse(line string) (map[string]interface{}, error) {
	return map[string]interface{}{"log": line}, nil
}

// Parses line as JSON
type JSONParser struct{}

func (p *JSONParser) Init() error { return nil }

func (p *JSONParser) Parse(line string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(line), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
