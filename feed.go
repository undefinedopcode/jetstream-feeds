package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Feed struct {
	ID               string          `hcl:"id,label"`
	Name             string          `hcl:"name"`
	PinnedURI        string          `hcl:"pinned_uri,optional"`
	Host             string          `hcl:"host,optional"`
	Port             int             `hcl:"port"`
	MatchExpr        string          `hcl:"match_expr,optional"`
	MatchAnalyzer    *AnalyzerConfig `hcl:"match_analyzer,block"`
	ForceExpr        string          `hcl:"force_expr,optional"`
	IncludeReplies   bool            `hcl:"include_replies,optional"`
	DB               string          `hcl:"database"`
	matcher          *regexp.Regexp
	forcer           *regexp.Regexp
	smatcher         *TextAnalyzer
	db               *gorm.DB
	ch               chan *Post
	PublishConfig    *PublishConfig `hcl:"publish,block"`
	ExclusionFilters []string       `hcl:"exclusion_filters,optional"`
	filters          map[string]*TextAnalyzer
	worker           *Worker
	r                *gin.Engine
}

type Pattern struct {
	Pattern    string  `hcl:"pattern,label"`
	Confidence float64 `hcl:"confidence"`
}

func (feed *Feed) GetLocalHostUrl(baseDid string) string {
	//http://localhost:6502/xrpc/app.bsky.feed.getFeedSkeleton?feed=at://did:plc:lbniuhsfce4bq2kqomky52px/app.bsky.feed.generator/neurodiversity
	url := fmt.Sprintf(
		"http://localhost:%d/xrpc/app.bsky.feed.getFeedSkeleton?feed=at://%s/app.bsky.feed.generator/%s",
		feed.Port,
		baseDid,
		feed.ID,
	)
	return url
}

func (feed *Feed) ShouldFilter(postText string) bool {
	for name, analyzer := range feed.filters {
		if score, filter := analyzer.Score(postText); filter {
			log.Info("Excluding due to sentiment score", "analyzer", name, "score", score, "text", postText)
			return true
		}
	}
	return false
}

func (feed *Feed) Matches(postText string, isReply bool) bool {
	if feed.ForceExpr != "" {
		if feed.forcer == nil {
			feed.forcer = regexp.MustCompile("(?i)" + feed.ForceExpr)
		}
		if feed.forcer.MatchString(postText) {
			return true
		}
	}
	if feed.MatchExpr != "" {
		if feed.matcher == nil {
			feed.matcher = regexp.MustCompile("(?i)" + feed.MatchExpr)
		}
		if feed.matcher.MatchString(postText) && (!isReply || (isReply && feed.IncludeReplies)) {
			return !feed.ShouldFilter(postText)
		}
		return false
	}
	if feed.MatchAnalyzer != nil {
		if feed.smatcher == nil {
			feed.smatcher = NewTextAnalyzer([]string{}, feed.MatchAnalyzer.Patterns, feed.MatchAnalyzer.Threshold, true)
		}
		if _, matches := feed.smatcher.Score(postText); !matches {
			return false
		}
	}

	return true
}

func (feed *Feed) StartProcessing(logger *log.Logger) {
	feed.worker = NewWorker(
		feed.ID+"-worker",
		feed.PostHandler,
		3, // number of retries
		3, // max concurrency
		false,
		dummyBackoffFunc,
		logger,
	)
	feed.worker.Start()
}

func (feed *Feed) Stop() {
	if feed.worker != nil {
		feed.worker.Stop()
	}
}

func (feed *Feed) PostHandler(job *WorkItem) (error, bool) {
	event := job.payload.(*models.Event)

	var post apibsky.FeedPost
	if err := json.Unmarshal(event.Commit.Record, &post); err != nil {
		return fmt.Errorf("failed to unmarshal post: %w", err), true // no retry if this fails
	}

	if feed.Matches(post.Text, post.Reply != nil) {
		uri := fmt.Sprintf("at://%s/%s/%s", event.Did, event.Commit.Collection, event.Commit.RKey)
		feed.worker.logger.Debug("Post match", "feed", feed.ID, "uri", uri)
		var reply_parent = ""
		var reply_root = ""
		// log.Printf("post time = %d", event.TimeUS / 1000)
		p := &Post{
			URI:       uri,
			CID:       event.Commit.CID,
			IndexedAt: fmt.Sprintf("%d", time.Now().UnixMilli()),
		}
		if post.Reply != nil {
			reply_parent = post.Reply.Parent.Uri
			reply_root = post.Reply.Root.Uri
			p.ReplyParent = &reply_parent
			p.ReplyRoot = &reply_root
		}
		// log.Printf("Writing post")
		feed.ch <- p
		if cfg.Debug {
			fmt.Printf(
				"[%s] %v |(%s)| %s\n",
				feed.ID,
				time.UnixMicro(event.TimeUS).Format("15:04:05"),
				event.Did,
				post.Text,
			)
		}
	}

	return nil, false
}
