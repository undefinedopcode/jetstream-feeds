package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bluesky-social/jetstream/pkg/client"
	"github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	serverAddr = "wss://jetstream2.us-east.bsky.network/subscribe"
	pdsHost    = "https://shimeji.us-east.host.bsky.network"
)

var cfg *Config
var fConfigName = flag.String("config", "feeds.hcl", "HCL Config file for feeds")

func main() {
	flag.Parse()

hupRentry:
	var needsHUP = false

	ctx, cancelFunc := context.WithCancel(context.Background())
	slog.SetDefault(slog.New(&charmSLogHandler{}))
	logger := log.Default()

	var err error

	cfg, err = readConfig(*fConfigName)
	if err != nil {
		log.Error("Failed to read config: %v", err)
		os.Exit(1)
	}
	log.Info("read config", "config", *cfg)
	// os.Exit(0)

	if *fPublishFeedName != "" {
		for _, feed := range cfg.Feeds {
			if feed.ID == *fPublishFeedName {
				// publish it
				fmt.Print("Enter your Bsky app password: ")
				password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					log.Fatalf("failed to read password: %v", err)
				}
				err = publishFeedGen(ctx, pdsHost, cfg.Owner, string(password), feed)
				if err != nil {
					log.Fatalf("failed to publish feed: %v", err)
				}
				return
			}
		}
		log.Error("Failed to find feed in config!")
		os.Exit(1)
	}

	config := client.DefaultClientConfig()
	config.WebsocketURL = serverAddr
	config.Compress = true

	h := &handler{
		seenSeqs: make(map[int64]struct{}),
	}

	scheduler := sequential.NewScheduler("jetstream_localdev", slog.Default(), h.HandleEvent)

	c, err := client.NewClient(config, slog.Default(), scheduler)
	if err != nil {
		log.Error("failed to create client: %v", err)
		os.Exit(1)
	}

	cursor := time.Now().Add(10 * -time.Minute).UnixMicro()

	// Every 5 seconds print the events read and bytes read and average event size
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		lctx, _ := context.WithCancel(ctx)
		for {
			select {
			case <-lctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				cursor = time.Now().Add(100 * -time.Millisecond).UnixMicro()
			}
		}
	}()

	// start up services for feeds
	for _, feed := range cfg.Feeds {
		httpctx, _ := context.WithCancel(ctx)
		startFeedService(httpctx, feed)
		dbctx, _ := context.WithCancel(ctx)
		postWriter(dbctx, feed)
		feed.StartProcessing(logger)
	}

	// attach signal handlers
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	go func() {
		signal := <-sigs
		log.Info("signal received", "signal", signal)
		switch signal {
		case syscall.SIGHUP:
			for _, feed := range cfg.Feeds {
				feed.Stop()
			}
			cancelFunc()
			needsHUP = true
		case syscall.SIGTERM, syscall.SIGINT:
			for _, feed := range cfg.Feeds {
				feed.Stop()
			}
			cancelFunc()
			log.Info("shutting down gracefully")
			time.Sleep(5 * time.Second)
		}
	}()

retry:

	// we begin reading from the jetstream here
	// it sometimes will fail and disconnect, so we reset the cursor
	// back one second and retry if it fails.
	// database is keyed on post uri being unique, so there won't
	// be dupes, and we reduce the risk of missing posts
	if err := c.ConnectAndRead(ctx, &cursor); err != nil {
		cursor = time.Now().Add(1000 * -time.Millisecond).UnixMicro()
		goto retry
	}

	// if the signal handler flagged that SIGHUP has been received, we'll all time
	// for context cancellations for running goroutines, and then restart
	if needsHUP {
		time.Sleep(5 * time.Second)
		log.Info("=================================================================================")
		log.Info(" A SERVICE HUP WAS REQUESTED")
		log.Info("---------------------------------------------------------------------------------")
		log.Info(" HTTP, Database and Feed Workers have been told to quit, and the service config")
		log.Info(" will be re-read from disk, so that changes can be applied.")
		log.Info("=================================================================================")
		goto hupRentry
	} else {
		time.Sleep(5 * time.Second)
	}

	// service finally stopped
	log.Info("shutdown")
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
			for _, feed := range cfg.Feeds {
				feed.worker.AddWork(event)
			}
		}
	}

	return nil
}
