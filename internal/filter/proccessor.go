package filter

import (
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_4"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
)

type FilterProcessor struct {
	ruleManager *RuleManager
}

func NewFilterProcessor(ruleManager *RuleManager) *FilterProcessor {
	return &FilterProcessor{
		ruleManager: ruleManager,
	}
}

// ProcessRequestForDSPV24 обрабатывает BidRequest v2.4 для DSP
func (fp *FilterProcessor) ProcessRequestForDSPV24(dspURL string, req *ortb_V2_4.BidRequest) *FilterResult {
	if req == nil {
		return &FilterResult{Allowed: false}
	}
	extractor := NewV24BidRequestExtractor(req)
	return fp.processRequestForDSP(dspURL, extractor)
}

// ProcessRequestForDSPV25 обрабатывает BidRequest v2.5 для DSP
func (fp *FilterProcessor) ProcessRequestForDSPV25(dspURL string, req *ortb_V2_5.BidRequest) *FilterResult {
	if req == nil {
		return &FilterResult{Allowed: false}
	}
	extractor := NewV25BidRequestExtractor(req)
	return fp.processRequestForDSP(dspURL, extractor)
}

// ProcessResponseForSPPV24 обрабатывает BidResponse v2.4 для SPP
func (fp *FilterProcessor) ProcessResponseForSPPV24(sppURL string, resp *ortb_V2_4.BidResponse) *FilterResult {
	if resp == nil {
		return &FilterResult{Allowed: false}
	}
	extractor := NewV24BidResponseExtractor(resp)
	return fp.processResponseForSPP(sppURL, extractor)
}

// ProcessResponseForSPPV25 обрабатывает BidResponse v2.5 для SPP
func (fp *FilterProcessor) ProcessResponseForSPPV25(sppURL string, resp *ortb_V2_5.BidResponse) *FilterResult {
	if resp == nil {
		return &FilterResult{Allowed: false}
	}
	extractor := NewV25BidResponseExtractor(resp)
	return fp.processResponseForSPP(sppURL, extractor)
}

// Универсальные методы для работы с экстракторами
func (fp *FilterProcessor) processRequestForDSP(dspURL string, extractor BidRequestExtractor) *FilterResult {
	rules := fp.ruleManager.GetRulesForDSP(dspURL)

	if len(rules) == 0 {
		return &FilterResult{
			Allowed: true,
		}
	}

	for _, rule := range rules {
		if !fp.evaluateRule(rule, extractor) {
			return &FilterResult{
				Allowed: false,
			}
		}
	}

	return &FilterResult{
		Allowed: true,
	}
}

func (fp *FilterProcessor) processResponseForSPP(sppURL string, extractor BidResponseExtractor) *FilterResult {
	rules := fp.ruleManager.GetRulesForSPP(sppURL)
	autoRules := fp.ruleManager.GetAutoRulesForSPP()

	allRules := append(rules, autoRules...)

	if len(allRules) == 0 {
		return &FilterResult{
			Allowed: true,
		}
	}

	for _, rule := range allRules {
		if !fp.evaluateResponseRule(rule, extractor) {
			return &FilterResult{
				Allowed: false,
			}
		}
	}

	return &FilterResult{
		Allowed: true,
	}
}

func (fp *FilterProcessor) evaluateRule(rule *FilterRule, extractor BidRequestExtractor) bool {
	fieldValue := extractor.ExtractFieldValue(rule.Field)
	return rule.Value.Compare(fieldValue)
}

func (fp *FilterProcessor) evaluateResponseRule(rule *FilterRule, extractor BidResponseExtractor) bool {
	fieldValue := extractor.ExtractFieldValue(rule.Field)
	return rule.Value.Compare(fieldValue)
}
