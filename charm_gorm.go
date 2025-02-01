package main

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"gorm.io/gorm/logger"
)

type gormCharmLogger struct{}

func (g *gormCharmLogger) LogMode(logger.LogLevel) logger.Interface {
	return &gormCharmLogger{}
}

func (g *gormCharmLogger) Info(ctx context.Context, msg string, arg ...interface{}) {
	log.Info(msg, arg...)
}

func (g *gormCharmLogger) Warn(ctx context.Context, msg string, arg ...interface{}) {
	log.Warn(msg, arg...)
}

func (g *gormCharmLogger) Error(ctx context.Context, msg string, arg ...interface{}) {
	log.Error(msg, arg...)
}

func (g *gormCharmLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	if err == nil {
		log.Debug("Gorm", "statement", sql, "rowaffected", rowsAffected)
	} else {
		log.Error("Gorm", "statement", sql, "error", err)
	}
}
