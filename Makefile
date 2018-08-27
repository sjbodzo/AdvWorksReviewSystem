# App specific for receiverd
RCV_DOCKERFILE	?= Dockerfile-rcv
RCV_APP_NAME 	?= product-review-rcv
RCV_APP_VERSION ?= 1.0.0
RCV_APP_PORT 	?= 8081
RCV_HOST_PORT	?= 8081
RCV_API_VERSION	?= v1 

# Redis specific parameters
REDIS_ENDPOINT	?= host.docker.internal
REDIS_PROC_Q	?= proc_queue
REDIS_REQ_Q		?= req_queue
REDIS_PORT		?= 6379

# Postgres database parameters
DB_ENDPOINT		?= host.docker.internal
DB_PORT 		?= 5432
DB_USER			?= postgres
DB_PW			?= postgres 

# App specific for approverd
APR_DOCKERFILE	?= Dockerfile-approve
APR_APP_NAME	?= product-review-approver
APR_APP_VERSION	?= 1.0.0

build-app:
	@echo "Building app..."
	docker-compose build

run-app: build-app
	@echo "Running app..."
	docker-compose up --force-recreate

docker-build-rcv:
	@echo "Building receiver version: $(RCV_APP_VERSION), api version: $(RCV_API_VERSION)"
	docker build -f $(RCV_DOCKERFILE) --name=$(RCV_APP_NAME):$(RCV_APP_VERSION)

# Note: this won't run without the database and redis cluster up!
docker-run-rcv:
	@echo "Running receiver to capture new product reviews..."
	docker run --entrypoint=./main \
				-it -p $(RCV_HOST_PORT):$(RCV_APP_PORT) \
				-apiPort=$(RCV_APP_PORT) \
				-apiVersion=$(RCV_API_VERSION) \
				-dbEndpoint=$(DB_ENDPOINT) \
				-dbPort=$(DB_PORT) \
				-dbUser=$(DB_USER) \
				-dbPw=$(DB_PW) \
				-redisEndpoint=$(REDIS_ENDPOINT) \
				-redisPort=$(REDIS_PORT)	

docker-build-approver:
	@echo "Building approver version: $(APR_APP_VERSION)"
	docker build -f $(APR_DOCKERFILE) --name=$(APR_APP_NAME):$(APR_APP_VERSION)

docker-run-approver:
	@echo "Running approver to approve new product reviews..."
	docker run --entrypoint=./main -it \
				-redisProcQueueName=$(REDIS_REQ_QUEUE) \
				-redisEndpoint=$(REDIS_PROC_QUEUE) \
				-redisEndpoint=$(REDIS_ENDPOINT) \
				-redisPort=$(REDIS_PORT)