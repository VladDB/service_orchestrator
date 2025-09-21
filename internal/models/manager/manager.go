package manager

import (
	"service_orchestrator/internal/components/globals"
	"sync"
)

// singleton for managing all units
type Manager struct {
	globalAction int
	units        map[int]globals.Unit
	mtx          sync.Mutex
}

var instance *Manager
var once sync.Once

func GetInstance() *Manager {
	once.Do(func() {
		instance = &Manager{
			units: map[int]globals.Unit{},
			mtx:   sync.Mutex{},
		}
	})
	return instance
}

func (m *Manager) AddUnit(settings globals.Unit) {}
func (m *Manager) RemoveUnit(id int)             {}

func (m *Manager) GetUnit(id int) globals.Unit {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	_, ok := m.units[id]

	if ok {
		return m.units[id]
	}
	return globals.Unit{}
}

func (m *Manager) UpdateUnitState(id, state int)                        {}
func (m *Manager) UpdateUnitSettings(id, settings globals.UnitSettings) {}

func (m *Manager) GetGlobalAction() int       { return 0 }
func (m *Manager) SetGlobalAction(action int) {}
