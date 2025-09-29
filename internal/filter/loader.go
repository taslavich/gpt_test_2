package filter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FileRuleLoader struct {
	ruleManager *RuleManager
	dspFilePath string
	sppFilePath string
	watcher     *fsnotify.Watcher
}

func NewFileRuleLoader(ruleManager *RuleManager, dspFilePath, sppFilePath string) *FileRuleLoader {
	return &FileRuleLoader{
		ruleManager: ruleManager,
		dspFilePath: dspFilePath,
		sppFilePath: sppFilePath,
	}
}

func (fl *FileRuleLoader) LoadDSPRules() error {
	data, err := os.ReadFile(fl.dspFilePath)
	if err != nil {
		return err
	}

	var config SimpleRuleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	if err := ValidateDSPConfig(&config); err != nil {
		return fmt.Errorf("DSP config validation failed: %v", err)
	}

	fl.ruleManager.ClearAllDSPRules()

	for dspID, dspSettings := range config.DSPs {
		seenRules := make(map[string]bool)

		for _, simpleRule := range dspSettings.Rules {
			ruleKey := fmt.Sprintf("%s_%s", simpleRule.Field, simpleRule.Condition)
			if seenRules[ruleKey] {
				continue
			}
			seenRules[ruleKey] = true

			rule, err := parseSimpleRule(simpleRule)
			if err != nil {
				return fmt.Errorf("Error parsing rule for DSP %s: %v", dspID, err) //5
			}

			if err := fl.ruleManager.AddRule(dspID, rule); err != nil {
				return fmt.Errorf("Error adding rule for DSP %s: %v", dspID, err) //6
			}
		}
	}

	return nil
}

func (fl *FileRuleLoader) LoadSPPRules() error {
	data, err := os.ReadFile(fl.sppFilePath)
	if err != nil {
		return err
	}

	var config SimpleRuleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	if err := ValidateSPPConfig(&config); err != nil {
		return fmt.Errorf("SPP config validation failed: %v", err)
	}

	fl.ruleManager.ClearAllSPPRules()

	for sppID, sppSettings := range config.SPPs {
		seenRules := make(map[string]bool)

		for _, simpleRule := range sppSettings.Rules {
			ruleKey := fmt.Sprintf("%s_%s", simpleRule.Field, simpleRule.Condition)
			if seenRules[ruleKey] {
				continue
			}
			seenRules[ruleKey] = true

			rule, err := parseSimpleRule(simpleRule)
			if err != nil {
				return fmt.Errorf("Error parsing rule for SPP %s: %v", sppID, err) //7
			}
			if err := fl.ruleManager.AddSPPRule(sppID, rule); err != nil {
				return fmt.Errorf("Error adding rule for SPP %s: %v", sppID, err) //8
			}
		}
	}

	return nil
}

func (fl *FileRuleLoader) Close() error {
	if fl.watcher != nil {
		return fl.watcher.Close()
	}
	return nil
}
