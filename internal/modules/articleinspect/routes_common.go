package articleinspect

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"gorm.io/gorm"
)

type uint64Param struct {
	Raw string
}

func (p uint64Param) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(uint64(0)))
}

func (p *uint64Param) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p uint64Param) Parse() (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(p.Raw), 10, 64)
}

type optionalUint64Param struct {
	Raw   string
	IsSet bool
}

func (p optionalUint64Param) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(uint64(0)))
}

func (p *optionalUint64Param) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *optionalUint64Param) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p optionalUint64Param) Value() (uint64, error) {
	if !p.IsSet {
		return 0, nil
	}
	return strconv.ParseUint(strings.TrimSpace(p.Raw), 10, 64)
}

type optionalIntParam struct {
	Raw   string
	IsSet bool
}

func (p optionalIntParam) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(int(0)))
}

func (p *optionalIntParam) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *optionalIntParam) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p optionalIntParam) Value() (int, error) {
	if !p.IsSet {
		return 0, nil
	}
	return strconv.Atoi(strings.TrimSpace(p.Raw))
}

type optionalBoolParam struct {
	Raw   string
	IsSet bool
}

func (p optionalBoolParam) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(true))
}

func (p *optionalBoolParam) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *optionalBoolParam) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p optionalBoolParam) Ptr() (*bool, error) {
	if !p.IsSet {
		return nil, nil
	}
	value, err := strconv.ParseBool(strings.TrimSpace(p.Raw))
	if err != nil {
		return nil, err
	}
	return &value, nil
}

type optionalInt8Param struct {
	Raw   string
	IsSet bool
}

func (p optionalInt8Param) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(int8(0)))
}

func (p *optionalInt8Param) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *optionalInt8Param) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p optionalInt8Param) Ptr() (*int8, error) {
	if !p.IsSet {
		return nil, nil
	}
	value, err := strconv.ParseInt(strings.TrimSpace(p.Raw), 10, 8)
	if err != nil {
		return nil, err
	}
	result := int8(value)
	return &result, nil
}

type optionalTimeParam struct {
	Raw   string
	IsSet bool
}

func (p optionalTimeParam) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(time.Time{}))
}

func (p *optionalTimeParam) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *optionalTimeParam) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p optionalTimeParam) Ptr() (*time.Time, error) {
	if !p.IsSet {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(p.Raw))
	if err != nil {
		return nil, err
	}
	return &value, nil
}

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

func articleInspectStatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrCategoryNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrKeywordNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrArticleNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrTaskNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrTaskDeleteNotAllowed):
		return http.StatusConflict, "task cannot be deleted"
	case errors.Is(err, ErrInvalidCategoryInput):
		return http.StatusBadRequest, "invalid category input"
	case errors.Is(err, ErrInvalidKeywordInput):
		return http.StatusBadRequest, "invalid keyword input"
	case errors.Is(err, ErrInvalidTaskInput):
		return http.StatusBadRequest, "invalid task input"
	case errors.Is(err, ErrInvalidResultQuery):
		return http.StatusBadRequest, "invalid result query"
	case errors.Is(err, ErrInvalidActionInput):
		return http.StatusBadRequest, "invalid action input"
	case errors.Is(err, ErrInvalidLogQuery):
		return http.StatusBadRequest, "invalid log query"
	case errors.Is(err, ErrInvalidArticleQuery):
		return http.StatusBadRequest, "invalid article query"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
