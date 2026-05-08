package articleinspect

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
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
