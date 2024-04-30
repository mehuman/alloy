package kafka_test

import (
	"testing"
	"time"

	"github.com/grafana/alloy/internal/component/otelcol"
	"github.com/grafana/alloy/internal/component/otelcol/exporter/kafka"
	"github.com/grafana/alloy/syntax"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalAlloy(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		expected map[string]interface{}
	}{
		{
			testName: "Defaults",
			cfg: `
				protocol_version = "2.0.0"
			`,
			expected: map[string]interface{}{
				"brokers":          []string{"localhost:9092"},
				"protocol_version": "2.0.0",
				"topic":            "otlp_spans",
				"encoding":         "otlp_proto",
				"timeout":          5 * time.Second,
				"authentication":   map[string]interface{}{},
				"metadata": map[string]interface{}{
					"full": true,
					"retry": map[string]interface{}{
						"max":     3,
						"backoff": 250 * time.Millisecond,
					},
				},
				"retry_on_failure": map[string]interface{}{
					//TODO: Add more parameters, which are not in the otel docs?
					"enabled":              true,
					"initial_interval":     5 * time.Second,
					"randomization_factor": 0.5,
					"multiplier":           1.5,
					"max_interval":         30 * time.Second,
					"max_elapsed_time":     5 * time.Minute,
				},
				"sending_queue": map[string]interface{}{
					"enabled":       true,
					"num_consumers": 10,
					"queue_size":    1000,
				},
				"producer": map[string]interface{}{
					"max_message_bytes":  1000000,
					"required_acks":      1,
					"compression":        "none",
					"flush_max_messages": 0,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var expected kafkaexporter.Config
			err := mapstructure.Decode(tc.expected, &expected)
			require.NoError(t, err)

			var args kafka.Arguments
			err = syntax.Unmarshal([]byte(tc.cfg), &args)
			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*kafkaexporter.Config)

			require.Equal(t, expected, *actual)
		})
	}
}

func TestDebugMetricsConfig(t *testing.T) {
	tests := []struct {
		testName string
		agentCfg string
		expected otelcol.DebugMetricsArguments
	}{
		{
			testName: "default",
			agentCfg: `
			protocol_version = "2.0.0"
			`,
			expected: otelcol.DebugMetricsArguments{
				DisableHighCardinalityMetrics: true,
			},
		},
		{
			testName: "explicit_false",
			agentCfg: `
			protocol_version = "2.0.0"
			debug_metrics {
				disable_high_cardinality_metrics = false
			}
			`,
			expected: otelcol.DebugMetricsArguments{
				DisableHighCardinalityMetrics: false,
			},
		},
		{
			testName: "explicit_true",
			agentCfg: `
			protocol_version = "2.0.0"
			debug_metrics {
				disable_high_cardinality_metrics = true
			}
			`,
			expected: otelcol.DebugMetricsArguments{
				DisableHighCardinalityMetrics: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args kafka.Arguments
			require.NoError(t, syntax.Unmarshal([]byte(tc.agentCfg), &args))
			_, err := args.Convert()
			require.NoError(t, err)

			require.Equal(t, tc.expected, args.DebugMetricsConfig())
		})
	}
}
