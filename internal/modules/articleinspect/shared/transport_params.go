package shared

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type Uint64Param struct {
	Raw string
}

func (p Uint64Param) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(uint64(0)))
}

func (p *Uint64Param) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p Uint64Param) Parse() (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(p.Raw), 10, 64)
}

type OptionalUint64Param struct {
	Raw   string
	IsSet bool
}

func (p OptionalUint64Param) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(uint64(0)))
}

func (p *OptionalUint64Param) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *OptionalUint64Param) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p OptionalUint64Param) Value() (uint64, error) {
	if !p.IsSet {
		return 0, nil
	}
	return strconv.ParseUint(strings.TrimSpace(p.Raw), 10, 64)
}

type OptionalIntParam struct {
	Raw   string
	IsSet bool
}

func (p OptionalIntParam) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(int(0)))
}

func (p *OptionalIntParam) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *OptionalIntParam) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p OptionalIntParam) Value() (int, error) {
	if !p.IsSet {
		return 0, nil
	}
	return strconv.Atoi(strings.TrimSpace(p.Raw))
}

type OptionalBoolParam struct {
	Raw   string
	IsSet bool
}

func (p OptionalBoolParam) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(true))
}

func (p *OptionalBoolParam) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *OptionalBoolParam) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p OptionalBoolParam) Ptr() (*bool, error) {
	if !p.IsSet {
		return nil, nil
	}
	value, err := strconv.ParseBool(strings.TrimSpace(p.Raw))
	if err != nil {
		return nil, err
	}
	return &value, nil
}

type OptionalInt8Param struct {
	Raw   string
	IsSet bool
}

func (p OptionalInt8Param) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(int8(0)))
}

func (p *OptionalInt8Param) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *OptionalInt8Param) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p OptionalInt8Param) Ptr() (*int8, error) {
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

type OptionalTimeParam struct {
	Raw   string
	IsSet bool
}

func (p OptionalTimeParam) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(time.Time{}))
}

func (p *OptionalTimeParam) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

func (p *OptionalTimeParam) OnParamSet(isSet bool, parsed any) {
	_ = parsed
	p.IsSet = isSet
}

func (p OptionalTimeParam) Ptr() (*time.Time, error) {
	if !p.IsSet {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(p.Raw))
	if err != nil {
		return nil, err
	}
	return &value, nil
}
