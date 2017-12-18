// This example declares a durable Exchange, an ephemeral (auto-delete) Queue,
// binds the Queue to the Exchange with a binding key, and consumes every
// message published to that Exchange with that routing key.
//
package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchange     = flag.String("exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue        = flag.String("queue", "test-queue", "Ephemeral AMQP queue name")
	bindingKey   = flag.String("key", "test-key", "AMQP binding key")
	consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	lifetime     = flag.Duration("lifetime", 0*time.Second, "lifetime of process before shutdown (0s=infinite)")
	verbose      = false
)

func init() {
	flag.Parse()
}

func main() {
	done := make(chan bool)
	// uris := []string{ "amqp://guest:guest@localhost:5672/", "amqp://guest:guest@localhost:5673/" }
	uri := "amqp://guest:guest@localhost:5672/"
	var cs [61440]*Consumer
	for batch := 0; batch < 30; batch++ {
		for i := 0; i < 2048; i++ {
			go func(idx0, idx1 int, u string) {
				c, err := NewConsumer(idx1, u, *exchange, *exchangeType, *queue, *bindingKey, *consumerTag)
				if err != nil {
					log.Printf("[WARNING] %d %d %s", idx0, idx1, err)
				}
				cs[idx1] = c
				if idx0 == 2047 {
					done <- true
				}
			}(i, (i + (batch * 2048)), uri)
		}
		log.Printf("[INFO] WAITING ON BATCH %d", batch)
		<-done
		log.Printf("[INFO] BATCH %d COMPLETE", batch)
		time.Sleep(1 * time.Second)
	}

	if *lifetime > 0 {
		log.Printf("running for %s", *lifetime)
		time.Sleep(*lifetime)
	} else {
		log.Printf("running forever")
		select {}
	}

	log.Printf("shutting down")

	for i := 0; i < 61440; i++ {
		c := cs[i]
		if err := c.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
	}
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func NewConsumer(idx int, amqpURI, exchange, exchangeType, queueName, key, ctag string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error
	
	if verbose {
		log.Printf("dialing %q", amqpURI)
	}

	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	if verbose {
		log.Printf("got Connection, getting Channel")
	}
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	if verbose {
		log.Printf("got Channel, declaring Exchange (%q)", exchange)
	}
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	q := fmt.Sprintf("%s-%d", queueName, idx)
	if verbose {
		log.Printf("declared Exchange, declaring Queue %q", q)
	}
	queue, err := c.channel.QueueDeclare(
		q,         // name of the queue
		false,     // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	if verbose {
		log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
			queue.Name, queue.Messages, queue.Consumers, key)
	}

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	if verbose {
		log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	}
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go handle(deliveries, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
