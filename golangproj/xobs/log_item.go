package xobs

import "time"

type Level int

const (
	LevelNotSet Level = 0
	LevelDbg    Level = 10
	LevelInfo   Level = 20
	LevelWarn   Level = 30
	LevelErr    Level = 40
)

type LogItem struct {
	Time   time.Time
	Level  Level
	front  [nAttrsInline]Attr
	nFront int
	back   []Attr
}

const nAttrsInline = 5
