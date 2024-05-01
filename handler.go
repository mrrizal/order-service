package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("order-handler")
)

func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "placeOrderHandler")
	defer span.End()

	if r.Method != http.MethodPost {
		errMessage := "Method not allowed"
		setErrorResponse(w, span, errMessage, http.StatusMethodNotAllowed)
		return
	}

	order, err := decodeOrderRequest(r)
	if err != nil {
		errMessage := "Invalid request body"
		setErrorResponse(w, span, errMessage, http.StatusBadRequest)
		return
	}

	// err = publishOrderToRabbitMQ(order)
	// if err != nil {
	// 	http.Error(w, "Failed to publish order", http.StatusInternalServerError)
	// 	return
	// }

	successMessage := fmt.Sprintf("Order received: %s", order.Food)
	setSuccessResponse(w, span, successMessage, http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
