package main

import (
	"context"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log"
)

type Post struct {
	URI         string `gorm:"primaryKey"`
	CID         string `gorm:"notNull"`
	ReplyParent *string
	ReplyRoot   *string
	IndexedAt   string
}

type SubState struct {
	Sservice string
	Cursor   int
}

func openDatabase(filename string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Post{}, &SubState{})
	return db, nil
}

const writeBufferSize = 1024 * 1024

func postWriter(ctx context.Context, cfg *FeedConfig) (chan *Post, error) {
	if cfg.db == nil {
		db, err := openDatabase(cfg.DB)
		if err != nil {
			return nil, err
		}
		cfg.db = db
	}
	log.Printf("Start consumer: %s", cfg.Name)
	cfg.ch = make(chan *Post, writeBufferSize)
	go func() {
		for p := range cfg.ch {
			if cfg.db != nil {
				//s := time.Now()
				// cfg.db.Delete(&Post{URI: p.URI})
				cfg.db.Create(p)
				//log.Printf("Written post to database in %s", time.Since(s))
			}
		}
	}()
	return cfg.ch, nil
}
