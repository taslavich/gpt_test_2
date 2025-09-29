package filter

import (
	"encoding/json"
)

type FieldType string

const (
	FieldBidFloor      FieldType = "bidfloor"
	FieldDeviceCountry FieldType = "device.geo.country"
	FieldAppID         FieldType = "app.id"
	FieldSiteID        FieldType = "site.id"
	FieldDeviceIP      FieldType = "device.ip"
	FieldBannerWidth   FieldType = "banner.w"
	FieldBannerHeight  FieldType = "banner.h"
	FieldBidNurl       FieldType = "bid.nurl"
	FieldBidBurl       FieldType = "bid.burl"

	FieldBidPrice    FieldType = "bid.price"
	FieldBidID       FieldType = "bid.id"
	FieldBidAdID     FieldType = "bid.adid"
	FieldBidImpID    FieldType = "bid.impid"
	FieldSeatBidSeat FieldType = "seatbid.seat"
	FieldBidArray    FieldType = "bid.array"
)

type ValueType string

const (
	ValueTypeInt    ValueType = "int"
	ValueTypeFloat  ValueType = "float"
	ValueTypeString ValueType = "string"
)

type ConditionType string

const (
	ConditionEqual        ConditionType = "equal"
	ConditionNotEqual     ConditionType = "not_equal"
	ConditionGreaterThan  ConditionType = "greater_than"
	ConditionGreaterEqual ConditionType = "greater_equal"
	ConditionLessThan     ConditionType = "less_than"
	ConditionLessEqual    ConditionType = "less_equal"
	ConditionBetween      ConditionType = "between"
	ConditionNotBetween   ConditionType = "not_between"
	ConditionExists       ConditionType = "exists"
)

type FieldValue struct {
	Type   ValueType
	Int    int
	Float  float64
	String string
}

func NewIntValue(value int) FieldValue {
	return FieldValue{Type: ValueTypeInt, Int: value}
}

func NewFloatValue(value float64) FieldValue {
	return FieldValue{Type: ValueTypeFloat, Float: value}
}

func NewStringValue(value string) FieldValue {
	return FieldValue{Type: ValueTypeString, String: value}
}

type ConditionValue interface {
	Type() ValueType
	Compare(value FieldValue) bool
}

type FilterRule struct {
	ID        string
	Field     FieldType
	Condition ConditionType
	Value     ConditionValue
}

type SimpleRuleConfig struct {
	Version string                 `json:"version"`
	DSPs    map[string]DSPSettings `json:"dsps"`
	SPPs    map[string]SPPSettings `json:"spps"`
}

type DSPSettings struct {
	Rules []SimpleRule `json:"rules"`
}

type SPPSettings struct {
	Rules []SimpleRule `json:"rules"`
}

type SimpleRule struct {
	Field     FieldType       `json:"field"`
	Condition ConditionType   `json:"condition"`
	ValueType ValueType       `json:"value_type"`
	Value     json.RawMessage `json:"value"`
}

type FilterResult struct {
	Allowed bool `json:"allowed"`
}

// BidRequestExtractor интерфейс для извлечения значений из BidRequest разных версий
type BidRequestExtractor interface {
	ExtractFieldValue(field FieldType) FieldValue
}

// BidResponseExtractor интерфейс для извлечения значений из BidResponse разных версий
type BidResponseExtractor interface {
	ExtractFieldValue(field FieldType) FieldValue
}
