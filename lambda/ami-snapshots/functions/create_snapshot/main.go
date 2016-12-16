package main

import (
	"github.com/alistanis/awstools/lambda"
	"github.com/alistanis/awstools/lambda/ami-snapshots/handlers"
)

func init() {}

func main() {
	lambda.HandleWithParams(handlers.RequiredParams, handlers.CreateSnapshot)
}
