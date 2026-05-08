package shared

import (
	"net/http"

	"github.com/dovetaill/article-sentinel/internal/api/response"
)

type OKEnvelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

type CreatedEnvelopeOutput struct {
	Status int `status:"201"`
	Body   response.Envelope
}

func SuccessOKEnvelope(status int, message string, data any) *OKEnvelopeOutput {
	return &OKEnvelopeOutput{Status: status, Body: response.OK(message, data)}
}

func SuccessCreatedEnvelope(message string, data any) *CreatedEnvelopeOutput {
	return &CreatedEnvelopeOutput{Status: http.StatusCreated, Body: response.OK(message, data)}
}

func FailureOKEnvelope(status int, message string) *OKEnvelopeOutput {
	return &OKEnvelopeOutput{Status: status, Body: response.Fail(status, message)}
}

func FailureCreatedEnvelope(status int, message string) *CreatedEnvelopeOutput {
	return &CreatedEnvelopeOutput{Status: status, Body: response.Fail(status, message)}
}
