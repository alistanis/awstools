package main

import (
	"github.com/alistanis/awstools/lambda"
	"github.com/alistanis/awstools/lambda/ami-snapshots/handlers"
)

func main() {
	lambda.HandleWithParams(nil, handlers.CreateSnapshot)
}
