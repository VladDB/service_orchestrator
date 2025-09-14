package controller

import (
	"log/slog"
	"service_orchestrator/internal/models/process"
)

// structure for controlling all units
type Controller struct {
	units []*process.Process
}

func (c *Controller) AddUnit(u *process.Process) {
	u.Id = len(c.units) + 1
	c.units = append(c.units, u)
	slog.Debug("Add new unit to the controller", "id", u.Id)
}

// starting all units
func (c *Controller) StarAllUnits() {}

// stopping all units
func (c *Controller) StopAllUnit() {}

// reload all units, read units config file and merge with current settings
func (c *Controller) ReloadUnits() {}

// main working function. Controlling all units in the thread
func Run() {}
