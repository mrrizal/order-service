package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/streadway/amqp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	QueueName = "orders"
	BrokerURL = "amqp://guest:guest@localhost:5672/"
)

var messageBrokerConn *amqp.Connection

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// Register handlers.
	// sync
	// orderHandler := OrderHandler{
	// 	service: sendOrderToKitchenSyncService{},
	// }

	// async
	messageBrokerConn, err := initRabbitMQ()
	if err != nil {
		panic(err)
	}
	orderHandler := OrderHandler{
		service: &sendOrderToKitchenAsyncService{conn: messageBrokerConn},
	}
	handleFunc("/order/", orderHandler.placeOrderHandler)

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	declareQueue()
	defer closeRabbitMQ(messageBrokerConn)

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func main() {
	fmt.Println("start oder service ...")
	if err := run(); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("\nshutdown order service ...")
}
