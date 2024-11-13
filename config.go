package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"gorm.io/gorm"
)

type Config struct {
	Owner string       `hcl:"feed_owner"`
	Base  string       `hcl:"feed_base"`
	Feeds []FeedConfig `hcl:"feed,block"`
	Debug bool `hcl:"debug,optional"`
}

type FeedConfig struct {
	ID string   `hcl:"id,label"`
	Name string `hcl:"name"`
	Port int    `hcl:"port"`
	MatchExpr string `hcl:"match_expr"`
	ForceExpr string `hcl:"force_expr,optional"`
	IncludeReplies bool `hcl:"include_replies,optional"`
	DB    string       `hcl:"database"`
	matcher *regexp.Regexp
	forcer *regexp.Regexp
	db *gorm.DB
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

func (fc *FeedConfig) Matches(postText string, isReply bool) bool {
	if fc.ForceExpr != "" {
		if fc.forcer == nil {
			fc.forcer = regexp.MustCompile("(?i)"+fc.ForceExpr)
		}
		if fc.forcer.MatchString(postText) {
			return true
		}
	}
	if fc.MatchExpr != "" {
		if fc.matcher == nil {
			fc.matcher = regexp.MustCompile("(?i)"+fc.MatchExpr)
		}
		if fc.matcher.MatchString(postText) {
			return true
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
	return config, nil
}
