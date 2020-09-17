package mqueue

import (
	"app/src/confopt/log"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type Producer interface {
	Produce(data PublicData)
	SetRoutingKey(key string)
}

type PublicData struct {
	Message  []byte
	RouteKey string
	Delay    int
}
type producer struct {
	sync.Mutex
	workerStatus

	channel         *amqp.Channel
	errorChannel    chan<- error
	exchange        string
	mandatory       bool
	immediate       bool
	options         Options
	publishChannel  chan PublicData
	routingKey      string
	shutdownChannel chan struct{}
}

func newProducer(channel *amqp.Channel, errorChannel chan<- error, config ProducerConfig) *producer {
	return &producer{
		channel:         channel,
		errorChannel:    errorChannel,
		exchange:        config.Exchange,
		options:         config.Options,
		mandatory:       config.Mandatory,
		immediate:       config.Immediate,
		publishChannel:  make(chan PublicData, config.BufferSize),
		routingKey:      config.RoutingKey,
		shutdownChannel: make(chan struct{}),
	}
}

func (producer *producer) worker() {
	producer.markAsRunning()

	for {
		select {
		case pubData := <-producer.publishChannel:
			err := producer.produce(pubData)
			if err != nil {
				producer.errorChannel <- err
			}
		case <-producer.shutdownChannel:
			producer.closeChannel()

			return
		}
	}
}

func (producer *producer) setChannel(channel *amqp.Channel) {
	producer.Lock()
	producer.channel = channel
	producer.Unlock()
}

func (producer *producer) SetRoutingKey(key string) {
	producer.Lock()
	producer.routingKey = key
	producer.Unlock()
}

func (producer *producer) closeChannel() {
	producer.Lock()
	if err := producer.channel.Close(); err != nil {
		producer.errorChannel <- err
	}
	producer.Unlock()
}

func (producer *producer) Produce(data PublicData) {

	tm := time.NewTimer(1 * time.Second)
	select {
	case producer.publishChannel <- data:
	case <-tm.C:
		log.Error("mq service error")
	}
}

func (producer *producer) produce(pubData PublicData) error {
	producer.Lock()
	defer producer.Unlock()

	var msg = amqp.Publishing{}
	if pub, ok := producer.options["contentType"]; ok {
		msg.ContentType = pub.(string)
	} else {
		msg.ContentType = "application/json"
	}
	if pub, ok := producer.options["deliveryMode"]; ok {
		msg.DeliveryMode = uint8(pub.(int))
	} else {
		msg.DeliveryMode = 1
	}

	msg.Body = pubData.Message
	msg.MessageId = uuid.NewV4().String()
	routingKey := producer.routingKey
	if len(pubData.RouteKey) > 0 {
		routingKey = pubData.RouteKey
	}

	if pubData.Delay > 0 {
		msg.Headers = amqp.Table{"delay": pubData.Delay * 1000}
	}

	log.Info(msg.MessageId + " " + producer.exchange + " " + routingKey + " " + string(msg.Body))
	return producer.channel.Publish(producer.exchange, routingKey, producer.mandatory, producer.immediate, msg)
}

func (producer *producer) Stop() {
	if producer.markAsStoppedIfCan() {
		producer.shutdownChannel <- struct{}{}
	}
}
