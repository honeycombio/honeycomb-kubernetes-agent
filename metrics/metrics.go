package metrics

import "time"

type ResourceMetrics struct {
	Resource *Resource
	Metrics
}

type Metrics map[string]*Metric

type MetricType int

const (
	MetricTypeInt   MetricType = 1
	MetricTypeFloat MetricType = 2
)

type Metric struct {
	Type       MetricType
	IsCounter  bool
	IntValue   *uint64
	FloatValue *float64
}

func (m *Metric) GetValue() float64 {
	if m.Type == MetricTypeInt {
		if m.IntValue == nil {
			return 0
		} else {
			return float64(*m.IntValue)
		}
	}

	if m.FloatValue == nil {
		return 0
	} else {
		return *m.FloatValue
	}
}

type CounterValue struct {
	Key       string
	Timestamp time.Time
	Value     *Metric
}
