package filter

type IntCondition struct {
	values []int
	cond   ConditionType
}

func (ic IntCondition) Type() ValueType { return ValueTypeInt }
func (ic IntCondition) Compare(fieldValue FieldValue) bool {
	if fieldValue.Type != ValueTypeInt {
		return false
	}

	switch ic.cond {
	case ConditionEqual:
		return fieldValue.Int == ic.values[0]
	case ConditionNotEqual:
		return fieldValue.Int != ic.values[0]
	case ConditionGreaterThan:
		return fieldValue.Int > ic.values[0]
	case ConditionGreaterEqual:
		return fieldValue.Int >= ic.values[0]
	case ConditionLessThan:
		return fieldValue.Int < ic.values[0]
	case ConditionLessEqual:
		return fieldValue.Int <= ic.values[0]
	case ConditionBetween:
		return fieldValue.Int >= ic.values[0] && fieldValue.Int <= ic.values[1]
	case ConditionNotBetween:
		return fieldValue.Int < ic.values[0] || fieldValue.Int > ic.values[1]
	case ConditionExists:
		return true
	default:
		return false
	}
}

type StringCondition struct {
	values []string
	cond   ConditionType
}

func (sc StringCondition) Type() ValueType { return ValueTypeString }
func (sc StringCondition) Compare(fieldValue FieldValue) bool {
	if fieldValue.Type != ValueTypeString {
		return false
	}

	switch sc.cond {
	case ConditionEqual:
		return fieldValue.String == sc.values[0]
	case ConditionNotEqual:
		return fieldValue.String != sc.values[0]
	case ConditionExists:
		return fieldValue.String != ""
	default:
		return false
	}
}

type FloatCondition struct {
	values []float64
	cond   ConditionType
}

func (fc FloatCondition) Type() ValueType { return ValueTypeFloat }
func (fc FloatCondition) Compare(fieldValue FieldValue) bool {
	if fieldValue.Type != ValueTypeFloat {
		return false
	}

	switch fc.cond {
	case ConditionEqual:
		return fieldValue.Float == fc.values[0]
	case ConditionNotEqual:
		return fieldValue.Float != fc.values[0]
	case ConditionGreaterThan:
		return fieldValue.Float > fc.values[0]
	case ConditionGreaterEqual:
		return fieldValue.Float >= fc.values[0]
	case ConditionLessThan:
		return fieldValue.Float < fc.values[0]
	case ConditionLessEqual:
		return fieldValue.Float <= fc.values[0]
	case ConditionBetween:
		return fieldValue.Float >= fc.values[0] && fieldValue.Float <= fc.values[1]
	case ConditionNotBetween:
		return fieldValue.Float < fc.values[0] || fieldValue.Float > fc.values[1]
	case ConditionExists:
		return true
	default:
		return false
	}
}
