package main

import (
	"github.com/charmbracelet/log"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	Owner     string            `hcl:"feed_owner"`
	Base      string            `hcl:"feed_base"`
	Feeds     []*Feed           `hcl:"feed,block"`
	Debug     bool              `hcl:"debug,optional"`
	Analyzers []*AnalyzerConfig `hcl:"analyzer,block"`
}

type PublishConfig struct {
	ServiceHost        string `hcl:"service_host,label"`
	ServiceIcon        string `hcl:"service_icon,optional"`
	ServiceShortName   string `hcl:"service_short_name,optional"`
	ServiceHumanName   string `hcl:"service_human_name,optional"`
	ServiceDescription string `hcl:"service_description,optional"`
	ServiceDID         string `hcl:"service_did"`
}

type AnalyzerConfig struct {
	ID         string             `hcl:"id,label"`
	Triggers   []string           `hcl:"triggers,optional"`
	Threshold  float64            `hcl:"threshold,optional"`
	Patterns   map[string]float64 `hcl:"patterns"`
	AnyTrigger bool               `hcl:"any_trigger,optional"`
}

func readConfig(filename string) (*Config, error) {
	var config = &Config{}
	err := hclsimple.DecodeFile(filename, nil, config)
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		return nil, err
	}
	log.Debug("Configuration is %#v", config)
	for _, fc := range config.Feeds {
		fc.filters = map[string]*TextAnalyzer{}
		for _, ac := range config.Analyzers {
			fc.filters[ac.ID] = NewTextAnalyzer(ac.Triggers, ac.Patterns, ac.Threshold, ac.AnyTrigger)
		}
	}
	return config, nil
}
