package articleinspect

import sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"

type okEnvelopeOutput = sharedpkg.OKEnvelopeOutput

type createdEnvelopeOutput = sharedpkg.CreatedEnvelopeOutput

func successOKEnvelope(status int, message string, data any) *okEnvelopeOutput {
	return sharedpkg.SuccessOKEnvelope(status, message, data)
}

func successCreatedEnvelope(message string, data any) *createdEnvelopeOutput {
	return sharedpkg.SuccessCreatedEnvelope(message, data)
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
	return sharedpkg.FailureOKEnvelope(status, message)
}

func failureCreatedEnvelope(status int, message string) *createdEnvelopeOutput {
	return sharedpkg.FailureCreatedEnvelope(status, message)
}
