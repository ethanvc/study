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

func (l Level) String() string {
	switch l {
	case LevelNotSet:
		return "NotSet"
	case LevelDbg:
		return "Dbg"
	case LevelInfo:
		return "Info"
	case LevelWarn:
		return "Warn"
	case LevelErr:
		return "Err"
	}
	return "Unknown"
}

type LogItem struct {
	Msg      string
	Time     time.Time
	Level    Level
	Position string
	front    [nAttrsInline]Attr
	nFront   int
	back     []Attr
	ObsCtx   *ObsContext
}

const nAttrsInline = 5
