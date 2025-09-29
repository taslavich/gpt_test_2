package filter

import (
	"encoding/json"
	"fmt"
)

func ValidateSimpleRule(simpleRule SimpleRule) error {

	if simpleRule.Field == "" {
		return fmt.Errorf("field is required")
	}
	if simpleRule.Condition == "" {
		return fmt.Errorf("condition is required")
	}
	if simpleRule.ValueType == "" {
		return fmt.Errorf("value_type is required")
	}

	if simpleRule.Condition == ConditionExists {
		switch simpleRule.ValueType {
		case ValueTypeString, ValueTypeInt, ValueTypeFloat:
			return nil
		default:
			return fmt.Errorf("unknown value type for exists condition: %s", simpleRule.ValueType)
		}
	}

	switch simpleRule.ValueType {
	case ValueTypeInt:
		return validateIntCondition(simpleRule.Value, simpleRule.Condition)
	case ValueTypeString:
		return validateStringCondition(simpleRule.Value, simpleRule.Condition)
	case ValueTypeFloat:
		return validateFloatCondition(simpleRule.Value, simpleRule.Condition)
	default:
		return fmt.Errorf("unknown value type: %s", simpleRule.ValueType)
	}
}

func validateIntCondition(value json.RawMessage, cond ConditionType) error {
	switch cond {
	case ConditionBetween, ConditionNotBetween:
		var values []int
		if err := json.Unmarshal(value, &values); err != nil {
			return fmt.Errorf("invalid int array: %v", err)
		}
		if len(values) != 2 {
			return fmt.Errorf("requires exactly 2 values, got %d", len(values))
		}
	default:
		var singleValue int
		if err := json.Unmarshal(value, &singleValue); err != nil {
			return fmt.Errorf("invalid int value: %v", err)
		}
	}
	return nil
}

func validateStringCondition(value json.RawMessage, cond ConditionType) error {
	switch cond {
	default:
		var singleValue string
		if err := json.Unmarshal(value, &singleValue); err != nil {
			return fmt.Errorf("invalid string value: %v", err)
		}
		if singleValue == "" {
			return fmt.Errorf("string value cannot be empty for condition %s", cond)
		}
	}
	return nil
}

func validateFloatCondition(value json.RawMessage, cond ConditionType) error {
	switch cond {
	case ConditionBetween, ConditionNotBetween:
		var values []float64
		if err := json.Unmarshal(value, &values); err != nil {
			return fmt.Errorf("invalid float array: %v", err)
		}
		if len(values) != 2 {
			return fmt.Errorf("requires exactly 2 values, got %d", len(values))
		}
	default:
		var singleValue float64
		if err := json.Unmarshal(value, &singleValue); err != nil {
			return fmt.Errorf("invalid float value: %v", err)
		}
	}
	return nil
}

func ValidateConfig(config *SimpleRuleConfig) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if (config.DSPs == nil || len(config.DSPs) == 0) &&
		(config.SPPs == nil || len(config.SPPs) == 0) {
		return fmt.Errorf("at least one DSP or SPP configuration is required")
	}

	if config.DSPs != nil && len(config.DSPs) > 0 {
		if err := ValidateDSPConfig(config); err != nil {
			return err
		}
	}

	if config.SPPs != nil && len(config.SPPs) > 0 {
		if err := ValidateSPPConfig(config); err != nil {
			return err
		}
	}

	return nil
}

func ValidateDSPConfig(config *SimpleRuleConfig) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.DSPs == nil || len(config.DSPs) == 0 {
		return fmt.Errorf("at least one DSP configuration is required")
	}

	for dspID, dspSettings := range config.DSPs {
		if dspID == "" {
			return fmt.Errorf("DSP ID cannot be empty")
		}

		seenRules := make(map[string]bool)
		for _, simpleRule := range dspSettings.Rules {
			ruleKey := fmt.Sprintf("%s_%s", simpleRule.Field, simpleRule.Condition)
			if seenRules[ruleKey] {
				return fmt.Errorf("duplicate rule for DSP %s: %s", dspID, ruleKey)
			}
			seenRules[ruleKey] = true

			if err := ValidateSimpleRule(simpleRule); err != nil {
				return fmt.Errorf("invalid rule for DSP %s: %v", dspID, err)
			}
		}
	}

	return nil
}

func ValidateSPPConfig(config *SimpleRuleConfig) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.SPPs == nil || len(config.SPPs) == 0 {
		return fmt.Errorf("at least one SPP configuration is required")
	}

	for sppID, sppSettings := range config.SPPs {
		if sppID == "" {
			return fmt.Errorf("SPP ID cannot be empty")
		}

		seenRules := make(map[string]bool)
		for _, simpleRule := range sppSettings.Rules {
			ruleKey := fmt.Sprintf("%s_%s", simpleRule.Field, simpleRule.Condition)
			if seenRules[ruleKey] {
				return fmt.Errorf("duplicate rule for SPP %s: %s", sppID, ruleKey)
			}
			seenRules[ruleKey] = true

			if err := ValidateSimpleRule(simpleRule); err != nil {
				return fmt.Errorf("invalid rule for SPP %s: %v", sppID, err)
			}
		}
	}

	return nil
}
