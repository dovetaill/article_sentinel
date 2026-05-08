package articleinspect

import (
	"net/http"

	"github.com/dovetaill/article-sentinel/internal/api/response"
)

type okEnvelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

type createdEnvelopeOutput struct {
	Status int `status:"201"`
	Body   response.Envelope
}

func successOKEnvelope(status int, message string, data any) *okEnvelopeOutput {
	return &okEnvelopeOutput{Status: status, Body: response.OK(message, data)}
}

func successCreatedEnvelope(message string, data any) *createdEnvelopeOutput {
	return &createdEnvelopeOutput{Status: http.StatusCreated, Body: response.OK(message, data)}
}

func failureOKFromError(err error) (*okEnvelopeOutput, error) {
	status, message := articleInspectStatusFromError(err)
	return failureOKEnvelope(status, message), nil
}

func failureCreatedFromError(err error) (*createdEnvelopeOutput, error) {
	status, message := articleInspectStatusFromError(err)
	return failureCreatedEnvelope(status, message), nil
}

func failureOKEnvelope(status int, message string) *okEnvelopeOutput {
	return &okEnvelopeOutput{Status: status, Body: response.Fail(status, message)}
}

func failureCreatedEnvelope(status int, message string) *createdEnvelopeOutput {
	return &createdEnvelopeOutput{Status: status, Body: response.Fail(status, message)}
}
