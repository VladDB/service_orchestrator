package manager

import (
	"sync"
	"time"
)

// unit state
const (
	Stop = iota
	Stopping
	Work
	Starting
	Error
)

// actions for units
const (
	NoAction = iota
	ToStop
	ToStart
	ToRestart
	ToStartAll
	ToStopAll
)

type UnitSettings struct {
	Name         string
	BinPath      string
	Args         []string
	StartTime    time.Time
	StopTime     time.Time
	AutoStart    bool
	RestartDelay uint
	State        int
}

type Unit struct {
	settings UnitSettings
	State    int
	Action   int
}

// singleton for managing all units
type Manager struct {
	globalAction  int
	unitsSettings map[int]Unit
	mtx           sync.Mutex
}

var instance *Manager
var once sync.Once

func GetInstance() *Manager {
	once.Do(func() {
		instance = &Manager{
			unitsSettings: map[int]Unit{},
			mtx:           sync.Mutex{},
		}
	})
	return instance
}

func (m *Manager) AddUnit(settings Unit)                        {}
func (m *Manager) RemoveUnit(id int)                            {}
func (m *Manager) UpdateUnitState(id, state int)                {}
func (m *Manager) UpdateUnitSettings(id, settings UnitSettings) {}

func (m *Manager) GetGlobalAction() int       { return 0 }
func (m *Manager) SetGlobalAction(action int) {}
