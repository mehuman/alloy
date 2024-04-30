// Package kafka provides an otelcol.exporter.kafka component.
package kafka

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/grafana/alloy/internal/component"
	"github.com/grafana/alloy/internal/component/otelcol"
	"github.com/grafana/alloy/internal/component/otelcol/exporter"
	alloy_kafka_receiver "github.com/grafana/alloy/internal/component/otelcol/receiver/kafka"
	"github.com/grafana/alloy/internal/featuregate"
	"github.com/grafana/alloy/syntax"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:      "otelcol.exporter.kafka",
		Stability: featuregate.StabilityPublicPreview,
		Args:      Arguments{},
		Exports:   otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := kafkaexporter.NewFactory()
			return exporter.New(opts, fact, args.(Arguments), exporter.TypeAll)
		},
	})
}

// Arguments configures the otelcol.exporter.kafka component.
type Arguments struct {
	Brokers         []string      `alloy:"brokers,attr,optional"`
	ProtocolVersion string        `alloy:"protocol_version,attr"`
	Topic           string        `alloy:"topic,attr,optional"`
	Encoding        string        `alloy:"encoding,attr,optional"`
	Timeout         time.Duration `alloy:"timeout,attr,optional"`

	Authentication alloy_kafka_receiver.AuthenticationArguments `alloy:"authentication,block,optional"`
	Metadata       alloy_kafka_receiver.MetadataArguments       `alloy:"metadata,block,optional"`
	Retry          otelcol.RetryArguments                       `alloy:"retry_on_failure,block,optional"`
	Queue          otelcol.QueueArguments                       `alloy:"sending_queue,block,optional"`
	Producer       Producer                                     `alloy:"producer,block,optional"`

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `alloy:"debug_metrics,block,optional"`
}

// Producer defines configuration for producer
type Producer struct {
	// Maximum message bytes the producer will accept to produce.
	MaxMessageBytes int `alloy:"max_message_bytes,attr,optional"`

	// RequiredAcks Number of acknowledgements required to assume that a message has been sent.
	// https://pkg.go.dev/github.com/IBM/sarama@v1.30.0#RequiredAcks
	// The options are:
	//   0 -> NoResponse.  doesn't send any response
	//   1 -> WaitForLocal. waits for only the local commit to succeed before responding ( default )
	//   -1 -> WaitForAll. waits for all in-sync replicas to commit before responding.
	RequiredAcks int `alloy:"required_acks,attr,optional"`

	// Compression Codec used to produce messages
	// https://pkg.go.dev/github.com/IBM/sarama@v1.30.0#CompressionCodec
	// The options are: 'none', 'gzip', 'snappy', 'lz4', and 'zstd'
	Compression string `alloy:"compression,attr,optional"`

	// The maximum number of messages the producer will send in a single
	// broker request. Defaults to 0 for unlimited. Similar to
	// `queue.buffering.max.messages` in the JVM producer.
	FlushMaxMessages int `alloy:"flush_max_messages,attr,optional"`
}

// Convert converts args into the upstream type.
func (args Producer) Convert() kafkaexporter.Producer {
	return kafkaexporter.Producer{
		MaxMessageBytes:  args.MaxMessageBytes,
		RequiredAcks:     sarama.RequiredAcks(args.RequiredAcks),
		Compression:      args.Compression,
		FlushMaxMessages: args.FlushMaxMessages,
	}
}

var (
	_ syntax.Validator   = (*Arguments)(nil)
	_ syntax.Defaulter   = (*Arguments)(nil)
	_ exporter.Arguments = (*Arguments)(nil)
)

// SetToDefault implements syntax.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = Arguments{
		Topic:    "otlp_spans",
		Encoding: "otlp_proto",
		Brokers:  []string{"localhost:9092"},
		Timeout:  5 * time.Second,
		Metadata: alloy_kafka_receiver.MetadataArguments{
			IncludeAllTopics: true,
			Retry: alloy_kafka_receiver.MetadataRetryArguments{
				MaxRetries: 3,
				Backoff:    250 * time.Millisecond,
			},
		},
		Producer: Producer{
			MaxMessageBytes:  1000000,
			RequiredAcks:     1,
			Compression:      "none",
			FlushMaxMessages: 0,
		},
	}
	args.Retry.SetToDefault()
	args.Queue.SetToDefault()
	args.DebugMetrics.SetToDefault()
}

// Validate implements syntax.Validator.
func (args *Arguments) Validate() error {
	//TODO: Implement this later
	return nil
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	input := make(map[string]interface{})
	input["auth"] = args.Authentication.Convert()

	var result kafkaexporter.Config
	err := mapstructure.Decode(input, &result)
	if err != nil {
		return nil, err
	}

	result.Brokers = args.Brokers
	result.ProtocolVersion = args.ProtocolVersion
	result.Topic = args.Topic
	result.Encoding = args.Encoding
	result.TimeoutSettings = exporterhelper.TimeoutSettings{
		Timeout: args.Timeout,
	}
	result.Metadata = args.Metadata.Convert()
	result.BackOffConfig = *args.Retry.Convert()
	result.QueueSettings = *args.Queue.Convert()
	result.Producer = args.Producer.Convert()

	return &result, nil
}

// Extensions implements exporter.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements exporter.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	return args.DebugMetrics
}
