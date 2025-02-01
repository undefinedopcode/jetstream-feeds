package main

import (
	"context"
	"time"

	"github.com/charmbracelet/log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
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
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{
		Logger: &gormCharmLogger{},
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Post{}, &SubState{})
	return db, nil
}

const writeBufferSize = 1024 * 1024

func postWriter(ctx context.Context, cfg *Feed) (chan *Post, error) {
	if cfg.db == nil {
		db, err := openDatabase(cfg.DB)
		if err != nil {
			return nil, err
		}
		cfg.db = db
	}
	log.Info("Starting database consumer", "feed", cfg.ID)
	cfg.ch = make(chan *Post, writeBufferSize)
	go func() {
		for {
			select {
			case <-ctx.Done():
				if cfg.db != nil {
					if rawdb, err := cfg.db.DB(); err == nil {
						log.Warn("Stopping database consumer", "feed", cfg.ID)
						rawdb.Close()
					}
				}
				return
			case p := <-cfg.ch:
				if cfg.db != nil {
					cfg.db.Create(p)
				}
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}()
	return cfg.ch, nil
}
