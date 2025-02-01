package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/labstack/echo/v4"
)

type PostRec struct {
	Post string `json:"post"`
}

type PostList struct {
	Cursor string    `json:"cursor"`
	Feed   []PostRec `json:"feed"`
}

func getDIDDoc(cfg *Feed) []byte {
	txt := fmt.Sprintf(`{
    "@context": [
        "https://www.w3.org/ns/did/v1"
    ],
    "id": "%s",
    "service": [
        {
            "id": "#bsky_fg",
            "type": "BskyFeedGenerator",
            "serviceEndpoint": "https://%s"
        }
    ]
}`, cfg.PublishConfig.ServiceDID, cfg.PublishConfig.ServiceHost)
	return []byte(txt)
}

func startFeedService(ctx context.Context, cfg *Feed) {
	r := echo.New()
	r.GET("/.well-known/atproto-did", func(c echo.Context) error {
		c.JSONBlob(http.StatusOK, getDIDDoc(cfg))
		return nil
	})
	r.GET("/.well-known/did.json", func(c echo.Context) error {
		c.JSONBlob(http.StatusOK, getDIDDoc(cfg))
		return nil
	})
	r.GET("/xrpc/app.bsky.feed.getFeedSkeleton", func(c echo.Context) error {
		feed := c.QueryParam("feed")
		limit := c.QueryParam("limit")
		cursor := c.QueryParam("cursor")
		log.Info("getFeedSkeleton Params", "feed", feed, "limit", limit, "cursor", cursor)
		if cfg.db == nil {
			if db, err := openDatabase(cfg.DB); err == nil {
				cfg.db = db
			}
		}
		if cfg.db != nil {
			if limit == "" {
				limit = "25"
			}
			ts := ""
			cid := ""
			iLimit, err := strconv.ParseInt(limit, 10, 32)
			if err != nil {
				c.String(400, fmt.Sprintf("Bad request: malformed limit param"))
				return nil
			}
			if cursor != "" && strings.Contains(cursor, "::") {
				parts := strings.SplitN(cursor, "::", 2)
				ts = parts[0]
				cid = parts[1]
				if ts == "" || cid == "" {
					c.String(401, fmt.Sprintf("Bad request: malformed cursor"))
					return nil
				}
				_, err = strconv.ParseInt(ts, 10, 64)
				if err != nil {
					c.String(402, fmt.Sprintf("Bad request: malformed cursor"))
					return nil
				}
			}
			var posts = []*Post{}
			if ts != "" && cid != "" {
				cfg.db.Limit(int(iLimit)).Where("c_id < ? and (indexed_at < ? or indexed_at = ?)", cid, ts, ts).Order("indexed_at desc, c_id desc").Find(&posts)
			} else {
				cfg.db.Limit(int(iLimit)).Order("indexed_at desc, c_id desc").Find(&posts)
			}
			// log.Printf("Got posts = %+v", posts)
			if len(posts) > 0 {
				last := posts[len(posts)-1]
				list := &PostList{
					Cursor: last.IndexedAt + "::" + last.CID,
					Feed:   []PostRec{},
				}
				if cid == "" && cfg.PinnedURI != "" {
					list.Feed = append(list.Feed, PostRec{cfg.PinnedURI})
				}
				for _, p := range posts {
					list.Feed = append(list.Feed, PostRec{p.URI})
				}
				c.JSON(http.StatusOK, list)
				return nil
			}
		}
		c.String(404, fmt.Sprintf("Posts not found"))
		return nil

	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	go func() {
		log.Info("Starting server...", "feed", cfg.ID, "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server:", "feed", cfg.ID, "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()
}
