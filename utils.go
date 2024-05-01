package main

import (
	"encoding/json"
	"net/http"
)

func decodeOrderRequest(r *http.Request) (Order, error) {
	var order Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		return Order{}, err
	}
	return order, nil
}
