package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PostRec struct {
	Post string `json:"post"`
}

type PostList struct {
	Cursor string `json:"cursor"`
	Feed []PostRec `json:"feed"`
}

func startFeedService(cfg *FeedConfig) {
	r := gin.Default()
	r.GET("/xrpc/app.bsky.feed.getFeedSkeleton", func(c *gin.Context) {
		feed := c.Query("feed")
		limit := c.Query("limit")
		cursor := c.Query("cursor")
		log.Printf("Params: feed=%s, limit=%s, cursor=%s", feed, limit, cursor)
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
				c.AbortWithError(400, fmt.Errorf("Bad request: malformed limit param"))
				return
			}
			if cursor != "" && strings.Contains(cursor, "::") {
				parts := strings.SplitN(cursor, "::", 2)
				ts = parts[0]
				cid = parts[1]
				if ts == "" || cid == "" {
					c.AbortWithError(400, fmt.Errorf("Bad request: malformed cursor"))
					return
				}
				_, err = strconv.ParseInt(ts, 10, 32)
				if err != nil {
					c.AbortWithError(400, fmt.Errorf("Bad request: malformed cursor"))
					return
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
				last := posts[0]
				list := &PostList{
					Cursor: last.IndexedAt+"::"+last.CID,
					Feed: []PostRec{
					},
				}
				for _, p := range posts {
					list.Feed = append(list.Feed, PostRec{p.URI})
				}
				c.JSON(http.StatusOK, list)
				return
			}
		}
		c.AbortWithError(404, fmt.Errorf("Posts not found"))

	})
	go r.Run(fmt.Sprintf(":%d", cfg.Port))
}
