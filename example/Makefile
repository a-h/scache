build:
	go get ./...
	env GOOS=linux go build -ldflags="-s -w" -o bin/api api/main.go

# Deploys to AWS (creates a Kinesis stream and DynamoDB database which cost money, so remember to shut it down).
deploy: build
	serverless deploy

# Deletes all created resources.
remove:
	serverless remove
