## Payment Gateway
### Architecture
![Architecture](https://zeze.nyc3.cdn.digitaloceanspaces.com/exinity/exinity.drawio.png)
### Services
The microservice has 4 main services
1) Api Service - Receives requests from clients and also process gateways callback
2) Payment Processor - Listen for transactions from kafka and forward the requests to the gateways
3) Callback Processor - Listen for transactions callbacks from gateways through Kafka and update transactions status
4) Callback Dispatcher - Forward transactions status back to the clients if callback url is set

### Design Doc
[The design doc can be found at the root of the project](Payment Gateway Design Doc.pdf)
### Running the microservice
``docker compose up --build``

### Using the rest api
OpenAPI specification can be found at the root of the project `api.yml`

### Running tests
Tests for gateway integrations and utils are provided
``go test -v ./...``
