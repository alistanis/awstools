package handlers

import (
	"encoding/json"

	"os"

	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
)

const (
	snapshotEnvVar = "SNAPSHOT_PATTERN"
	roleEnvVar     = "IAM_ROLE"
	handlerEnvVar  = "HANDLER"
)

var (
	snapshotPattern = os.Getenv(snapshotEnvVar)
	role            = os.Getenv(roleEnvVar)
	handler         = os.Getenv(handlerEnvVar)

	RequiredParams = map[string]string{
		snapshotEnvVar: snapshotPattern,
		roleEnvVar:     role,
		handlerEnvVar:  handler,
	}
)

func CreateSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	return "Hello, World!", nil
}

func CheckSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	return "", nil
}

func DeleteSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	return "", nil
}
