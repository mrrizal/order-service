package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/propagation"
)

type sendOrderToKitchen interface {
	send(ctx context.Context, order Order) error
}

type sendOrderToKitchenSyncService struct {
}

type sendOrderToKitchenAsyncService struct {
	conn *amqp.Connection
}

func (s sendOrderToKitchenSyncService) send(ctx context.Context, order Order) error {
	ctx, span := tracer.Start(ctx, "send-sync")
	defer span.End()

	url := "http://localhost:8081/cook/"

	jsonData, err := json.Marshal(order)
	if err != nil {
		return err
	}

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))

	client := &http.Client{}
	resp, err := client.Do(r)
	log.Printf("status code: %d\n", resp.StatusCode)
	defer resp.Body.Close()

	if err != nil {
		return err
	}
	return nil
}

type AMQPHeaderCarrier struct {
	Headers amqp.Table
}

func (a AMQPHeaderCarrier) Set(key, value string) {
	a.Headers[key] = value
}

func (a AMQPHeaderCarrier) Get(key string) string {
	v, ok := a.Headers[key]
	if !ok {
		return ""
	}
	return v.(string)
}

func (a AMQPHeaderCarrier) Keys() []string {
	i := 0
	r := make([]string, len(a.Headers))

	for k := range a.Headers {
		r[i] = k
		i++
	}

	return r
}

func (s *sendOrderToKitchenAsyncService) send(ctx context.Context, order Order) error {
	ctx, span := tracer.Start(ctx, "send-async")
	defer span.End()

	if s.conn == nil && s.conn.IsClosed() {
		return errors.New("amqp not connected")
	}

	ch, err := s.conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	headersCarrier := amqp.Table{}
	propagator.Inject(ctx, AMQPHeaderCarrier{headersCarrier})

	ch.Publish("", QueueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
		Headers:     headersCarrier,
	})
	return nil
}
