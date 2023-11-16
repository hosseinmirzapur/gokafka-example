build-consumer:
	go build -o bin/consumer ./cmd/consumer/consumer.go

build-producer:
	go build -o bin/producer ./cmd/producer/producer.go

consumer: build-consumer
	./bin/consumer

producer: build-producer
	./bin/producer