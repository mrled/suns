package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event map[string]interface{}) error {
	fmt.Println("Hello from scheduled Lambda!")
	return nil
}

func main() {
	lambda.Start(handler)
}
