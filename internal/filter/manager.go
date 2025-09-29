package filter

import (
	"sync"
)

type RuleManager struct {
	dspRules map[string]map[string]*FilterRule
	sppRules map[string]map[string]*FilterRule
	mu       sync.RWMutex
}

func NewRuleManager() *RuleManager {
	return &RuleManager{
		dspRules: make(map[string]map[string]*FilterRule),
		sppRules: make(map[string]map[string]*FilterRule),
	}
}

func (rm *RuleManager) AddRule(dspID string, rule *FilterRule) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.dspRules[dspID]; !exists {
		rm.dspRules[dspID] = make(map[string]*FilterRule)
	}

	rm.dspRules[dspID][rule.ID] = rule
	return nil
}

func (rm *RuleManager) AddSPPRule(sppID string, rule *FilterRule) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.sppRules[sppID]; !exists {
		rm.sppRules[sppID] = make(map[string]*FilterRule)
	}

	rm.sppRules[sppID][rule.ID] = rule
	return nil
}

func (rm *RuleManager) GetAutoRulesForSPP() []*FilterRule {
	autoRules := []*FilterRule{
		{
			ID:        "auto_nurl_exists",
			Field:     FieldBidNurl,
			Condition: ConditionExists,
			Value:     StringCondition{cond: ConditionExists, values: []string{}},
		},
		{
			ID:        "auto_burl_exists",
			Field:     FieldBidBurl,
			Condition: ConditionExists,
			Value:     StringCondition{cond: ConditionExists, values: []string{}},
		},
	}
	return autoRules
}

func (rm *RuleManager) GetRulesForDSP(dspURL string) []*FilterRule {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*FilterRule, 0)
	if dspRules, exists := rm.dspRules[dspURL]; exists {
		for _, rule := range dspRules {
			rules = append(rules, rule)
		}
	}
	return rules
}

func (rm *RuleManager) GetRulesForSPP(sppURL string) []*FilterRule {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*FilterRule, 0)
	if sppRules, exists := rm.sppRules[sppURL]; exists {
		for _, rule := range sppRules {
			rules = append(rules, rule)
		}
	}
	return rules
}

func (rm *RuleManager) GetAllDSPRules() map[string][]*FilterRule {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string][]*FilterRule)
	for dspID, rules := range rm.dspRules {
		dspRules := make([]*FilterRule, 0, len(rules))
		for _, rule := range rules {
			dspRules = append(dspRules, rule)
		}
		result[dspID] = dspRules
	}
	return result
}

func (rm *RuleManager) GetAllSPPRules() map[string][]*FilterRule {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string][]*FilterRule)
	for sppID, rules := range rm.sppRules {
		sppRules := make([]*FilterRule, 0, len(rules))
		for _, rule := range rules {
			sppRules = append(sppRules, rule)
		}
		result[sppID] = sppRules
	}
	return result
}

func (rm *RuleManager) RemoveRule(dspID, ruleID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if dspRules, exists := rm.dspRules[dspID]; exists {
		delete(dspRules, ruleID)
	}
}

func (rm *RuleManager) RemoveSPPRule(sppID, ruleID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if sppRules, exists := rm.sppRules[sppID]; exists {
		delete(sppRules, ruleID)
	}
}

func (rm *RuleManager) ClearRules() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.dspRules = make(map[string]map[string]*FilterRule)
	rm.sppRules = make(map[string]map[string]*FilterRule)
}

func (rm *RuleManager) ClearAllDSPRules() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.dspRules = make(map[string]map[string]*FilterRule)
}

func (rm *RuleManager) ClearAllSPPRules() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.sppRules = make(map[string]map[string]*FilterRule)
}

func (rm *RuleManager) ClearDSPRules(dspID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.dspRules, dspID)
}

func (rm *RuleManager) ClearSPPRules(sppID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.sppRules, sppID)
}

func (rm *RuleManager) DSPExists(dspID string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	_, exists := rm.dspRules[dspID]
	return exists
}

func (rm *RuleManager) SPPExists(sppID string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	_, exists := rm.sppRules[sppID]
	return exists
}
