package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello from webhook Lambda!",
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
