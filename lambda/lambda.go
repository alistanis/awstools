package lambda

import (
	"encoding/json"

	"fmt"

	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
)

func HandleWithParams(requiredParams map[string]string, handle func(evt json.RawMessage, ctx *runtime.Context) (interface{}, error)) {
	f := func(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
		for k, p := range requiredParams {
			if p == "" {
				return nil, fmt.Errorf("%s was not found and is a required environment variable.", k)
			}
		}
		return handle(evt, ctx)
	}
	runtime.HandleFunc(f)
}
