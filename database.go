package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Post struct {
  URI string `gorm:"primaryKey"` 
  CID string `gorm:"notNull"`
  ReplyParent *string
  ReplyRoot *string 
  IndexedAt string
}

type SubState struct {
  Sservice string
  Cursor  int
}

func openDatabase(filename string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Post{}, &SubState{})
	return db, nil
}

func init() {
	_, _ = openDatabase("feed.db")
}
