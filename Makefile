run:
	go mod tidy
	export OTEL_RESOURCE_ATTRIBUTES="service.name=order-service,service.version=0.1.0"
	go run .
