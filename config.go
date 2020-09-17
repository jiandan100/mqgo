package mqueue

import "time"

type MQConfig struct {
	Queues map[string]*Config `yaml:"mq"`
}

type Config struct {
	DSN            string        `mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	Ali            bool          `mapstructure:"ali" json:"ali" yaml:"ali"`
	Ak             string        `mapstructure:"ak" json:"ak" yaml:"ak"`
	Sk             string        `mapstructure:"sk" json:"sk" yaml:"sk"`
	AliInstanceId  string        `mapstructure:"aliInstanceId" json:"aliInstanceId" yaml:"aliInstanceId"`
	ReconnectDelay time.Duration `mapstructure:"reconnect_delay" json:"reconnect_delay" yaml:"reconnect_delay"`
	Exchanges      Exchanges     `mapstructure:"exchanges" json:"exchanges" yaml:"exchanges"`
	Queues         Queues        `mapstructure:"queues" json:"queues" yaml:"queues"`
	Producers      Producers     `mapstructure:"producers" json:"producers" yaml:"producers"`
	Consumers      Consumers     `mapstructure:"consumers" json:"consumers" yaml:"consumers"`
}

type (
	Exchanges []ExchangeConfig
	Queues    []QueueConfig
	Consumers []ConsumerConfig
	Producers []ProducerConfig

	ExchangeConfig struct {
		Name    string  `json:"name" yaml:"name"`
		Type    string  `json:"type" yaml:"type"`
		Options Options `json:"options" yaml:"options"`
	}
	QueueConfig struct {
		Exchange       string  `json:"exchange" yaml:"exchange"`
		Name           string  `json:"name" yaml:"name"`
		RoutingKey     string  `json:"routing_key" yaml:"routing_key"`
		BindingOptions Options `json:"binding_options" yaml:"binding_options"`
		Options        Options `json:"options" yaml:"options"`
	}
	ConsumerConfig struct {
		Name          string  `json:"name" yaml:"name"`
		Queue         string  `json:"queue" yaml:"queue"`
		Workers       int     `json:"workers" yaml:"workers"`
		Options       Options `json:"options" yaml:"options"`
		PrefetchCount int     `json:"prefetch_count" yaml:"prefetch_count"`
		PrefetchSize  int     `json:"prefetch_size" yaml:"prefetch_size"`
	}
	ProducerConfig struct {
		BufferSize int     `json:"buffer_size" yaml:"buffer_size"`
		Exchange   string  `json:"exchange" yaml:"exchange"`
		Name       string  `json:"name" yaml:"name"`
		RoutingKey string  `json:"routing_key" yaml:"routing_key"`
		Options    Options `json:"options" yaml:"options"`
		Mandatory  bool    `json:"mandatory" yaml:"mandatory"`
		Immediate  bool    `json:"immediate" yaml:"immediate"`
	}
)

func (cfg *Config) normalize() {
	for _, exchange := range cfg.Exchanges {
		exchange.Options.normalizeKeys()
		exchange.Options.buildArgs()
	}
	for _, queue := range cfg.Queues {
		queue.Options.normalizeKeys()
		queue.BindingOptions.normalizeKeys()
		queue.Options.buildArgs()
		queue.BindingOptions.buildArgs()
	}
	for _, producer := range cfg.Producers {
		producer.Options.normalizeKeys()
	}
	for _, consumer := range cfg.Consumers {
		consumer.Options.normalizeKeys()
		consumer.Options.buildArgs()
	}
}
