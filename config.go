package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"gorm.io/gorm"
)

type Config struct {
	Owner     string            `hcl:"feed_owner"`
	Base      string            `hcl:"feed_base"`
	Feeds     []*FeedConfig     `hcl:"feed,block"`
	Debug     bool              `hcl:"debug,optional"`
	Analyzers []*AnalyzerConfig `hcl:"analyzer,block"`
}

type FeedConfig struct {
	ID               string `hcl:"id,label"`
	Name             string `hcl:"name"`
	PinnedURI        string `hcl:"pinned_uri,optional"`
	Host             string `hcl:"host,optional"`
	Port             int    `hcl:"port"`
	MatchExpr        string `hcl:"match_expr"`
	ForceExpr        string `hcl:"force_expr,optional"`
	IncludeReplies   bool   `hcl:"include_replies,optional"`
	DB               string `hcl:"database"`
	matcher          *regexp.Regexp
	forcer           *regexp.Regexp
	db               *gorm.DB
	ch               chan *Post
	PublishConfig    *PublishConfig `hcl:"publish,block"`
	ExclusionFilters []string       `hcl:"exclusion_filters,optional"`
	filters          map[string]*TextAnalyzer
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
	ID        string             `hcl:"id,label"`
	Triggers  []string           `hcl:"triggers,optional"`
	Threshold float64            `hcl:"threshold,optional"`
	Patterns  map[string]float64 `hcl:"patterns"`
	AnyTrigger bool `hcl:"any_trigger,optional"`
}

type Pattern struct {
	Pattern    string  `hcl:"pattern,label"`
	Confidence float64 `hcl:"confidence"`
}

func (fc *FeedConfig) GetLocalHostUrl(baseDid string) string {
	//http://localhost:6502/xrpc/app.bsky.feed.getFeedSkeleton?feed=at://did:plc:lbniuhsfce4bq2kqomky52px/app.bsky.feed.generator/neurodiversity
	url := fmt.Sprintf(
		"http://localhost:%d/xrpc/app.bsky.feed.getFeedSkeleton?feed=at://%s/app.bsky.feed.generator/%s",
		fc.Port,
		baseDid,
		fc.ID,
	)
	return url
}

func (fc *FeedConfig) ShouldFilter(postText string) bool {
	for name, analyzer := range fc.filters {
		if score, filter := analyzer.Score(postText); filter {
			log.Printf("[%s] Excluding due to sentiment score of > %f: %s", name, score, postText)
			return true
		}
	}
	return false
}

func (fc *FeedConfig) Matches(postText string, isReply bool) bool {
	if fc.ForceExpr != "" {
		if fc.forcer == nil {
			fc.forcer = regexp.MustCompile("(?i)" + fc.ForceExpr)
		}
		if fc.forcer.MatchString(postText) {
			return true
		}
	}
	if fc.MatchExpr != "" {
		if fc.matcher == nil {
			fc.matcher = regexp.MustCompile("(?i)" + fc.MatchExpr)
		}
		if fc.matcher.MatchString(postText) && (!isReply || (isReply && fc.IncludeReplies)) {
			return !fc.ShouldFilter(postText)
		}
		return false
	}

	return true
}

func readConfig(filename string) (*Config, error) {
	var config = &Config{}
	err := hclsimple.DecodeFile(filename, nil, config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	log.Printf("Configuration is %#v", config)
	for _, fc := range config.Feeds {
		fc.filters = map[string]*TextAnalyzer{}
		for _, ac := range config.Analyzers {
			fc.filters[ac.ID] = NewTextAnalyzer(ac.Triggers, ac.Patterns, ac.Threshold, ac.AnyTrigger)
		}
	}
	return config, nil
}
