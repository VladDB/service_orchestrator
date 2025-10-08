package globals

import (
	"sync/atomic"
	"time"
)

var ProcessRunning atomic.Bool

var BinPath string

var HttpPort int

var LogLevel int

// unit state
const (
	State_Stop = iota
	State_Stopping
	State_Work
	State_Starting
	State_Failed
	State_Timeout
)

// actions for units
const (
	Act_NoAction = iota
	Act_Stop
	Act_Start
	Act_Restart
)

// actions for all units
const (
	Act_Global_NoAction = iota
	Act_Global_StartAll
	Act_Global_StopAll
	Act_Global_ReloadAll
)

var ProcSystemStates = map[string]int{
	"R": State_Work,
	"S": State_Work,
	"D": State_Work,
	"T": State_Stop,
	"Z": State_Stop,
	"t": State_Stop,
	"X": State_Failed,
	"x": State_Failed,
	"W": State_Failed,
}

type UnitSettings struct {
	Name         string   // unit's name
	Cmd          string   // path to exec file
	Args         []string // args for the exec file
	UseRestart   bool     // flag for restarting unit if it failed
	RestartDelay uint     // delay before restart of the unit
	AutoStart    bool     // start with start main process
}

type Unit struct {
	Settings  UnitSettings
	State     int
	Action    int
	StartTime time.Time
	StopTime  time.Time
}

func CreateUnit(settings UnitSettings) *Unit {
	var unit Unit
	unit.Settings = settings
	unit.State = State_Stop
	unit.Action = Act_NoAction
	return &unit
}
