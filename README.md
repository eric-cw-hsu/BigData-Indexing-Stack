# INFO7255 Application

## Overview

This project is organized as two Go microservices:

- **API Service (`cmd/api-service`)**  
  A RESTful web service built with Gin that handles CRUD operations for “plans”.  
  • Data Storage: MongoDB  
  • Message Queue: RabbitMQ  
  • Auth: Google OAuth middleware  
- **Elasticsearch Service (`cmd/elasticsearch-service`)**  
  A background consumer that listens to plan events on RabbitMQ and updates Elasticsearch.  
  • Health-check endpoint at `/health` 

## Project Structure
```
├── cmd/
│ ├── api-service/ # REST API microservice
│ │ └── main.go
│ └── elasticsearch-service/ # Elasticsearch consumer microservice
│ └── main.go
└── internal/
  ├── api/
  │ ├── config/ # service‐specific config structs
  │ ├── handlers/ # HTTP handlers (plan_handler.go, …)
  │ ├── repositories/ # data access (plan_repository.go)
  │ ├── services/ # business logic (plan_service.go)
  │ ├── routes/ # router setup (router.go)
  │ ├── schema/ # JSON Schema files & loader
  │ └── utils/ # shared error types, ETag, etc.
  ├── elasticsearch/
  │ ├── client.go # ES client wrapper
  │ ├── processor.go # message → index/delete logic
  │ ├── service.go # Start() orchestration
  │ └── mappings/ # index mappings JSON
  └── objectstore/ # graph node extraction & Mongo storage
    ├── extractor.go
    ├── repository.go
    ├── retriever.go
    └── merger.go
```

## Configuration

Place your YAML files under `config/` in project root:

### config/api-service.yaml
```yaml
server:
  port: 8080
mongo:
  uri: "mongodb://localhost:27017"
  database: "plans_db"
rabbitmq:
  uri: "amqp://guest:guest@localhost:5672/"
  queue: "plans"
oauth:
  google_client_id: "<your-google-client-id>"
```

### config/elasticsearch-service.yaml
```yaml
elastic_search:
  addr: "http://localhost:9200"
  username: ""
  password: ""
  index: "plans"
  health_checker_port: "8081"
rabbitmq:
  uri: "amqp://guest:guest@localhost:5672/"
  queue: "plans"
```

## Running the Services

### API Service

1. Ensure MongoDB and RabbitMQ are running.  
2. From project root, run:
   ```bash
   go run cmd/api-service/main.go \
     --config config/api-service.yaml
   ```
3. The API listens on the port defined under `server.port`.

### Elasticsearch Service

1. Ensure RabbitMQ and Elasticsearch are running.  
2. From project root, run:
   ```bash
   go run cmd/elasticsearch-service/main.go \
     --config config/elasticsearch-service.yaml
   ```
3. Health-check is served on the port in `elastic_search.health_checker_port`.