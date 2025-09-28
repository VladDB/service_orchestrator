package manager

import (
	"fmt"
	"log/slog"
	"service_orchestrator/internal/components/globals"
	"sync"
	"time"
)

// singleton for managing all units
type Manager struct {
	globalAction int
	units        map[int]*globals.Unit
	mtx          sync.Mutex
}

var instance *Manager
var once sync.Once

func GetInstance() *Manager {
	once.Do(func() {
		instance = &Manager{
			units: make(map[int]*globals.Unit),
			mtx:   sync.Mutex{},
		}
	})
	return instance
}

func (m *Manager) AddUnit(settings globals.UnitSettings) int {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var newId int
	unitsLen := len(m.units)
	if unitsLen == 0 {
		newId = 1
	} else {
		newId = unitsLen + 1
	}
	m.units[newId] = globals.CreateUnit(settings)
	slog.Debug("Add new unit to manager",
		"name", settings.Name,
		"cmd", settings.Cmd,
		"args", settings.Args)

	return newId
}

func (m *Manager) RemoveUnit(id int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	delete(m.units, id)
}

func (m *Manager) GetUnit(id int) (globals.Unit, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if u, ok := m.units[id]; ok {
		return *u, nil
	}

	return globals.Unit{}, fmt.Errorf("unit with id %d not found", id)
}

func (m *Manager) UpdateUnitState(id, state int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if u, ok := m.units[id]; ok {
		switch state {
		case globals.State_Stop:
			u.StopTime = time.Now()
		case globals.State_Work:
			u.StartTime = time.Now()
		}
		u.State = state
	}
}

func (m *Manager) SetUnitAction(id, act int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if u, ok := m.units[id]; ok {
		u.Action = act
	}
}

func (m *Manager) SetUnitStopTime(id int, stopTime time.Time) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if u, ok := m.units[id]; ok {
		u.StopTime = stopTime
	}
}

func (m *Manager) UpdateUnitSettings(id int, settings globals.UnitSettings) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if u, ok := m.units[id]; ok {
		u.Settings = settings
	}
}

func (m *Manager) GetGlobalAction() int {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.globalAction
}

func (m *Manager) SetGlobalAction(action int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.globalAction = action
}
