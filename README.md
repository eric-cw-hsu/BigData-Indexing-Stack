# INFO7255 Application

## Overview

This project combines several components to build a robust application with the following key features:
- **Web API:** A RESTful service built with Gin that handles CRUD operations for plans.
- **Data Storage:** Uses Redis as a NoSQL database.
- **Message Queue:** Utilizes RabbitMQ to publish asynchronous events during plan creation, update, and deletion.
- **Search Indexing:** Integrates with Elasticsearch for indexing and searching plan data.
- **Microservices:** A dedicated microservice, the Elastic Consumer, consumes RabbitMQ messages and updates Elasticsearch accordingly.

## Project Structure
```
├── config.yaml # Application configuration file
├── README.md # This file 
├── configs/ # Configuration package (SetupConfig, etc.) 
├── controllers/ # HTTP controllers (e.g., PlanController) 
├── middlewares/ # Middleware for logging and error handling 
├── microservices/ # Independent microservices (e.g., Elastic Consumer) 
├── repositories/ # Data repository implementations (Redis) 
├── routes/ # HTTP route definitions 
├── services/ # Business logic and external service integrations (Plan, ElasticSearch) 
├── utils/
└── queue/ # RabbitMQ integration package
```

## Key Components

- **Web API**
  - Implements CRUD operations for "plans".
  - Uses Gin for routing and integrates with Redis to store plan data.
  - Publishes events to a RabbitMQ queue for asynchronous processing.

- **RabbitMQ Integration**
  - The `queue` package provides a `RabbitMQQueue` which handles connection, channel, and queue declaration.
  - The PlanService uses this module to publish create, update, and delete events.

- **Elasticsearch Integration**
  - The `services/elastic_search.go` package encapsulates integration with Elasticsearch for indexing and deletion operations.
  - This service is leveraged by the dedicated microservice to update search indexes.

- **Elastic Consumer Microservice**
  - Located under `/microservices/elastic_search`, it consumes messages from RabbitMQ and updates Elasticsearch accordingly.
  - Exposes a simple HTTP endpoint (e.g., `/health`) for health checks.

## Configuration

The project uses a YAML configuration file (`config.yaml`) in the project root. A sample configuration is shown below:

```yaml
ENV: ""
PORT: 
REDIS:
  ADDR: ""
  DB: 
OAUTH2:
  ISSUER: "https://accounts.google.com"
  CLIENT_ID: ""
RABBITMQ:
  ADDR: ""
ELASTICSEARCH:
  ADDR: ""
  USERNAME: ""
  PASSWORD: ""
  HEALTH_CHECKER_PORT: ""
```

## Running the Application
Web API
1. Ensure Redis, RabbitMQ, and Elasticsearch services are running.
2. Build and run the web application:
```bash
$ go run main.go
```
3. The HTTP API will run on the port specified in the configuration (e.g., 8080).
   
Elastic Consumer Microservice
1. Navigate to the microservices directory:
   ```
   cd microservices/elastic_search
   ```
2. Build and run the microservice:
   ```
   $ go run main.go
   ```
3. The microservice provides a health check endpoint at `/health` (default port e.g., 8081).

## Logging and Error Handling
- Logging: The application uses Logrus for structured logging.
- Error Handling: Critical configuration and dependency errors cause the application to exit. Background workers log errors without impacting primary APIs.
