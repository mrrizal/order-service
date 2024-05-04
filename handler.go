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

type OrderHandler struct {
	service sendOrderToKitchen
}

func (o OrderHandler) placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "placeOrderHandler")
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

	successMessage := fmt.Sprintf("Order received: %s", order.Food)
	err = o.service.send(ctx, order)
	if err != nil {
		setErrorResponse(w, span, err.Error(), http.StatusInternalServerError)
		return
	}
	setSuccessResponse(w, span, successMessage, http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
