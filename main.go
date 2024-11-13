package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/client"
	"github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
)

const (
	serverAddr = "wss://jetstream.atproto.tools/subscribe"
)

var cfg *Config

func main() {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})))
	logger := slog.Default()

	var err error

	cfg, err = readConfig("feeds.hcl")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	log.Printf("read config: %+v", *cfg)
	// os.Exit(0)

	config := client.DefaultClientConfig()
	config.WebsocketURL = serverAddr
	config.Compress = true

	h := &handler{
		seenSeqs: make(map[int64]struct{}),
	}

	scheduler := sequential.NewScheduler("jetstream_localdev", logger, h.HandleEvent)

	c, err := client.NewClient(config, logger, scheduler)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	cursor := time.Now().Add(1 * -time.Minute).UnixMicro()

	// Every 5 seconds print the events read and bytes read and average event size
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				cursor = time.Now().Add(100 * -time.Millisecond).UnixMicro()
				eventsRead := c.EventsRead.Load()
				bytesRead := c.BytesRead.Load()
				avgEventSize := bytesRead / eventsRead
				logger.Info("stats", "events_read", eventsRead, "bytes_read", bytesRead, "avg_event_size", avgEventSize, "cursor", cursor)
			}
		}
	}()

	for _, feed := range cfg.Feeds {
		startFeedService(&feed)
	}

	retry:

	if err := c.ConnectAndRead(ctx, &cursor); err != nil {
		cursor = time.Now().Add(100 * -time.Millisecond).UnixMicro()
		goto retry
	}

	slog.Info("shutdown")
}

type handler struct {
	seenSeqs  map[int64]struct{}
	highwater int64
}

func (h *handler) HandleEvent(ctx context.Context, event *models.Event) error {
	// Unmarshal the record if there is one
	if event.Commit != nil && (event.Commit.Operation == models.CommitOperationCreate || event.Commit.Operation == models.CommitOperationUpdate) {
		switch event.Commit.Collection {
		case "app.bsky.feed.post":
			var post apibsky.FeedPost
			if err := json.Unmarshal(event.Commit.Record, &post); err != nil {
				return fmt.Errorf("failed to unmarshal post: %w", err)
			}

			for _, feed := range cfg.Feeds {
				if feed.Matches(post.Text, post.Reply != nil) {
					uri := fmt.Sprintf("at://%s/%s/%s", event.Did, event.Commit.Collection, event.Commit.RKey)
					var reply_parent = ""
					var reply_root = ""
					p := &Post{
						URI: uri,
						CID: event.Commit.CID,
						IndexedAt: fmt.Sprintf("%d", time.Now().UnixMilli()),
					}
					if post.Reply != nil {
						reply_parent = post.Reply.Parent.Uri
						reply_root = post.Reply.Root.Uri
						p.ReplyParent = &reply_parent
						p.ReplyRoot = &reply_root
					}
					if feed.db == nil {
						if db, err := openDatabase(feed.DB); err == nil {
							feed.db = db
						}
					}
					if feed.db != nil {
						feed.db.Delete(&Post{URI: p.URI})
						feed.db.Create(p)
					}
					if cfg.Debug {
						fmt.Printf(
							"[%s] %v |(%s)| %s\n",
							feed.ID,
							time.UnixMicro(event.TimeUS).Local().Format("15:04:05"),
							event.Did,
							post.Text,
						)
					}
				}				
			}
		}
	}

	return nil
}
