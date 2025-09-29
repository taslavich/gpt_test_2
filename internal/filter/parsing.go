package filter

import (
	"encoding/json"
	"fmt"
)

func parseSimpleRule(simpleRule SimpleRule) (*FilterRule, error) {
	if err := ValidateSimpleRule(simpleRule); err != nil {
		return nil, err
	}

	rule := &FilterRule{
		ID:        generateRuleID(simpleRule),
		Field:     simpleRule.Field,
		Condition: simpleRule.Condition,
	}

	if simpleRule.Condition == ConditionExists {
		switch simpleRule.ValueType {
		case ValueTypeString:
			rule.Value = StringCondition{cond: simpleRule.Condition, values: []string{}}
		case ValueTypeInt:
			rule.Value = IntCondition{cond: simpleRule.Condition, values: []int{}}
		case ValueTypeFloat:
			rule.Value = FloatCondition{cond: simpleRule.Condition, values: []float64{}}
		}
		return rule, nil
	}

	var err error
	switch simpleRule.ValueType {
	case ValueTypeInt:
		rule.Value, err = parseIntCondition(simpleRule.Value, simpleRule.Condition)
	case ValueTypeString:
		rule.Value, err = parseStringCondition(simpleRule.Value, simpleRule.Condition)
	case ValueTypeFloat:
		rule.Value, err = parseFloatCondition(simpleRule.Value, simpleRule.Condition)
	}

	if err != nil {
		return nil, err
	}

	return rule, nil
}

func parseIntCondition(value json.RawMessage, cond ConditionType) (IntCondition, error) {
	var intCond IntCondition
	intCond.cond = cond

	switch cond {
	case ConditionBetween, ConditionNotBetween:
		var values []int
		if err := json.Unmarshal(value, &values); err != nil {
			return intCond, fmt.Errorf("invalid int array: %v", err)
		}
		if len(values) != 2 {
			return intCond, fmt.Errorf("requires exactly 2 values, got %d", len(values))
		}
		intCond.values = values
	default:
		var singleValue int
		if err := json.Unmarshal(value, &singleValue); err != nil {
			return intCond, fmt.Errorf("invalid int value: %v", err)
		}
		intCond.values = []int{singleValue}
	}

	return intCond, nil
}

func parseStringCondition(value json.RawMessage, cond ConditionType) (StringCondition, error) {
	var strCond StringCondition
	strCond.cond = cond

	switch cond {
	default:
		var singleValue string
		if err := json.Unmarshal(value, &singleValue); err != nil {
			return strCond, fmt.Errorf("invalid string value: %v", err)
		}
		if singleValue == "" {
			return strCond, fmt.Errorf("string value cannot be empty for condition %s", cond)
		}
		strCond.values = []string{singleValue}
	}

	return strCond, nil
}

func parseFloatCondition(value json.RawMessage, cond ConditionType) (FloatCondition, error) {
	var floatCond FloatCondition
	floatCond.cond = cond

	switch cond {
	case ConditionBetween, ConditionNotBetween:
		var values []float64
		if err := json.Unmarshal(value, &values); err != nil {
			return floatCond, fmt.Errorf("invalid float array: %v", err)
		}
		if len(values) != 2 {
			return floatCond, fmt.Errorf("requires exactly 2 values, got %d", len(values))
		}
		floatCond.values = values
	default:
		var singleValue float64
		if err := json.Unmarshal(value, &singleValue); err != nil {
			return floatCond, fmt.Errorf("invalid float value: %v", err)
		}
		floatCond.values = []float64{singleValue}
	}

	return floatCond, nil
}

func generateRuleID(rule SimpleRule) string {
	return string(rule.Field) + "_" + string(rule.Condition)
}
