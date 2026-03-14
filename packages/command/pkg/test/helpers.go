package test

import (
	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
)

// CommandResponseGetter is implemented by response types that have GetOutput() and GetResult() methods.
type CommandResponseGetter interface {
	GetOutput() *commandv1beta1.CommandOutput
	GetResult() *commandv1beta1.CommandResult
}

// CollectedResponse holds the aggregated data from command responses.
type CollectedResponse struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int32
	Reason   commandv1beta1.CommandResult_Reason
}

// CollectCommandResponses collects stdout, stderr, exit code, and reason from a slice of responses.
func CollectCommandResponses[T CommandResponseGetter](responses []T) CollectedResponse {
	var result CollectedResponse
	for _, resp := range responses {
		if out := resp.GetOutput(); out != nil {
			result.Stdout = append(result.Stdout, out.Stdout...)
			result.Stderr = append(result.Stderr, out.Stderr...)
		}
		if res := resp.GetResult(); res != nil {
			result.ExitCode = res.ExitCode
			result.Reason = res.Reason
		}
	}
	return result
}
