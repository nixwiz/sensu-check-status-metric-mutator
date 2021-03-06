package main

import (
	"fmt"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu-community/sensu-plugin-sdk/templates"
	"github.com/sensu/sensu-go/types"
)

// Config represents the mutator plugin config.
type Config struct {
	sensu.PluginConfig
	MetricNameTemplate string
}

var (
	mutatorConfig = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-check-status-metric-mutator",
			Short:    "Sensu Check Status Metric Mutator",
			Keyspace: "sensu.io/plugins/sensu-check-status-metric-mutator/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "metric-name-template",
			Env:       "METRIC_NAME_TEMPLATE",
			Argument:  "metric-name-template",
			Shorthand: "t",
			Default:   "{{.Check.Name}}.status",
			Usage:     "Template for naming the metric point for the check status",
			Value:     &mutatorConfig.MetricNameTemplate,
		},
	}
)

func main() {
	mutator := sensu.NewGoMutator(&mutatorConfig.PluginConfig, options, checkArgs, executeMutator)
	mutator.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(mutatorConfig.MetricNameTemplate) == 0 {
		return fmt.Errorf("--MetricNameTemplate or METRIC_NAME_TEMPLATE environment variable is required")
	}
	return nil
}

func executeMutator(event *types.Event) (*types.Event, error) {
	if !event.HasCheck() {
		return &types.Event{}, fmt.Errorf("Event does not have a check defined.")
	}
	metricName, err := templates.EvalTemplate("metricName", mutatorConfig.MetricNameTemplate, event)
	if err != nil {
		return &types.Event{}, fmt.Errorf("Failed to evalutate template: %v", err)
	}

	// Possible TODO:  replace any spaces periods from the templated metricName

	// This really shouldn't happen if a metrics handler is defined, but just in case.
	if !event.HasMetrics() {
		event.Metrics = new(types.Metrics)
	}

	// Provide some extra information in the tags
	mt := make([]*types.MetricTag, 0)
	mt = append(mt, &types.MetricTag{Name: "entity", Value: event.Entity.Name})
	mt = append(mt, &types.MetricTag{Name: "check", Value: event.Check.Name})
	mt = append(mt, &types.MetricTag{Name: "state", Value: event.Check.State})
	mt = append(mt, &types.MetricTag{Name: "occurrences", Value: fmt.Sprintf("%d", event.Check.Occurrences)})
	mt = append(mt, &types.MetricTag{Name: "occurrences_watermark", Value: fmt.Sprintf("%d", event.Check.OccurrencesWatermark)})

	mp := &types.MetricPoint{
		Name:      metricName,
		Value:     float64(event.Check.Status),
		Timestamp: event.Timestamp,
		Tags:      mt,
	}
	event.Metrics.Points = append(event.Metrics.Points, mp)
	return event, nil
}
