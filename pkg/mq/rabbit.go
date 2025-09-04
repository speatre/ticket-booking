package mq

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"ticket-booking/pkg/config"
)

// Publisher interface để dễ mock
type Publisher interface {
	Publish(routingKey string, msg interface{}) error
}

// Consumer interface để dễ mock
type Consumer interface {
	Consume(queue string, handler func([]byte) error) error
}

// --- Implementation Publisher ---
type AMQPPublisher struct {
	ch       *amqp091.Channel
	exchange string
}

func NewPublisher(ch *amqp091.Channel, exchange string) *AMQPPublisher {
	if err := ch.ExchangeDeclare(
		exchange,
		config.DefaultExchangeType,
		true,  // durable
		false, // auto-delete
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchange declare: %v", err)
	}
	return &AMQPPublisher{ch: ch, exchange: exchange}
}

func (p *AMQPPublisher) Publish(routingKey string, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.ch.Publish(
		p.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// --- Implementation Consumer ---
type AMQPConsumer struct {
	ch       *amqp091.Channel
	exchange string
	queue    string
	key      string
}

func NewConsumer(ch *amqp091.Channel, exchange, queue, bindingKey string) *AMQPConsumer {
	_, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue declare: %v", err)
	}

	if err := ch.QueueBind(queue, bindingKey, exchange, false, nil); err != nil {
		log.Fatalf("queue bind: %v", err)
	}

	return &AMQPConsumer{ch: ch, exchange: exchange, queue: queue, key: bindingKey}
}

func (c *AMQPConsumer) Consume(queue string, handler func([]byte) error) error {
	msgs, err := c.ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("consume error: %v", err)
			}
		}
	}()
	return nil
}

// NoOpPublisher implements Publisher interface for when RabbitMQ is disabled
type NoOpPublisher struct{}

func (p *NoOpPublisher) Publish(routingKey string, msg interface{}) error {
	// No-op implementation - just log and return success
	log.Printf("NoOpPublisher: would publish to %s: %+v", routingKey, msg)
	return nil
}

// MustDial kết nối RabbitMQ, panic nếu lỗi
func MustDial(url string) *amqp091.Connection {
	conn, err := amqp091.Dial(url)
	if err != nil {
		log.Fatalf("❌ rabbitmq dial: %v", err)
	}
	return conn
}
