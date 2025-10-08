package controller

import (
	"log/slog"
	"service_orchestrator/internal/components/cache"
	"service_orchestrator/internal/components/configuration"
	"service_orchestrator/internal/components/globals"
	"service_orchestrator/internal/models/manager"
	"service_orchestrator/internal/models/process"
	"strings"
	"sync"
	"time"
)

// structure for controlling all units
type Controller struct {
	processes []*process.Process
}

func (c *Controller) AddUnit(settings globals.UnitSettings) {
	var p process.Process
	p.Id = manager.GetInstance().AddUnit(settings)
	c.processes = append(c.processes, &p)
	slog.Debug("Add new unit to the controller", "id", p.Id)
}

// starting all units
func (c *Controller) startAllUnits(checkAutoStart bool) {
	slog.Debug("Starting all processes")
	var wg sync.WaitGroup
	unitManager := manager.GetInstance()
	for _, proc := range c.processes {
		unit, err := unitManager.GetUnit(proc.Id)
		if err != nil {
			slog.Error("Failed to get unit while starting all", "id", proc.Id, "error", err.Error())
			continue
		}
		if checkAutoStart && !unit.Settings.AutoStart {
			continue
		}
		wg.Go(func() {
			proc.Start()
		})
	}
	wg.Wait()
	slog.Debug("All processes were started")
}

// stopping all units
func (c *Controller) stopAllUnit() {
	slog.Debug("Stopping all processes")
	var wg sync.WaitGroup
	for _, proc := range c.processes {
		wg.Go(func() {
			proc.Stop()
		})
	}
	wg.Wait()
	slog.Debug("All processes were stopped")
}

// reload all units, read units config file and merge with current settings
// restarting updated units
func (c *Controller) reloadUnits() {
	slog.Info("Reload units configuration")

	unitManager := manager.GetInstance()
	// prepare current settings
	var currentSettings []struct {
		proc    *process.Process
		setting globals.UnitSettings
	}

	for _, proc := range c.processes {
		procUnit, err := unitManager.GetUnit(proc.Id)
		if err != nil {
			slog.Error("Failed to get unit while reloading", "id", proc.Id, "error", err.Error())
			continue
		}
		currentSettings = append(currentSettings, struct {
			proc    *process.Process
			setting globals.UnitSettings
		}{proc: proc, setting: procUnit.Settings})
	}

	// get new settings
	newSettings, err := configuration.ReadConfiguration(nil)
	if err != nil {
		slog.Error("Failed to read new configuration", "error", err.Error())
		return
	}

	// prepare map for quick access
	newSettingsMap := make(map[string]globals.UnitSettings)
	for _, newUnit := range *newSettings {
		newSettingsMap[newUnit.Name] = newUnit
	}

	var wg sync.WaitGroup
	// compare new and old settings
	for _, curr := range currentSettings {
		if newUnit, ok := newSettingsMap[curr.setting.Name]; ok {
			// compare cmd and args
			if newUnit.Cmd != curr.setting.Cmd ||
				!equalStringSlices(newUnit.Args, curr.setting.Args) {
				slog.Info("Unit settings changed, reload it", "name", curr.setting.Name)
				wg.Go(func() {
					curr.proc.Stop()
					unitManager.UpdateUnitSettings(curr.proc.Id, newUnit)
					curr.proc.Start()
				})
			} else {
				slog.Debug("Unit settings are unchanged", "name", curr.setting.Name)
			}
		} else {
			slog.Warn("Unit was deleted", "name", curr.setting.Name)
			curr.proc.Stop()
			unitManager.RemoveUnit(curr.proc.Id)

			// delete process
			for i, p := range c.processes {
				if p.Id == curr.proc.Id {
					c.processes = append(c.processes[:i], c.processes[i+1:]...)
					break
				}
			}
		}
		// delete from new settings
		delete(newSettingsMap, curr.setting.Name)
	}

	// create new units if we have
	for _, newUnit := range newSettingsMap {
		c.AddUnit(newUnit)
	}

	// starting units with auto start
	for _, proc := range c.processes {
		proc.UpdateStatus()
		unit, err := unitManager.GetUnit(proc.Id)
		if err != nil {
			slog.Error("Failed to get unit while starting", "id", proc.Id, "error", err.Error())
			continue
		}
		if unit.State != globals.State_Work && unit.Settings.AutoStart {
			wg.Go(func() {
				proc.Start()
			})
		}
	}

	// waiting for all units
	wg.Wait()
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range b {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// main working function. Controlling all units in the thread
func (c *Controller) Run() {
	checkDelay := 5
	slog.Info("Run controller loop")

	unitManager := manager.GetInstance()

	// check cache processes
	cacheUnits := cache.LoadUnits()
	for _, uCache := range cacheUnits.Units {
		findUnit := false
		for _, proc := range c.processes {
			uSettings, err := unitManager.GetUnit(proc.Id)

			if err != nil && uSettings.Settings.Name == uCache.Name {
				// if find process than take pid
				proc.Pid = uCache.Pid
				slog.Debug("Process found in cache", "name", uCache.Name, "pid", uCache.Pid)

				// if cmd or args changed than stop process
				if !equalStringSlices(uSettings.Settings.Args, strings.Split(uCache.Args, " ")) || uSettings.Settings.Cmd != uCache.Cmd {
					slog.Info("Process found in cache, but cmd or args changed, stop it", "name", uCache.Name)
					proc.Stop()
				}
				findUnit = true
				break
			}
		}
		if findUnit {
			continue
		}

		// if don't find unit, than stop process from cache
		process.KillProcessByPid(uCache.Pid)
		cache.DeleteUnitFromFile(uCache.Name)
	} // for _, uCache := range cachUnits

	// clear cache slice
	cacheUnits.Units = nil

	// start all with auto start
	c.startAllUnits(true)

	time.Sleep(time.Duration(checkDelay) * time.Second)

	for globals.ProcessRunning.Load() {
		// check global actions
		globalAct := unitManager.GetGlobalAction()
		if globalAct != globals.Act_Global_NoAction {
			switch globalAct {
			case globals.Act_Global_StartAll:
				c.startAllUnits(false)
			case globals.Act_Global_StopAll:
				c.stopAllUnit()
			case globals.Act_Global_ReloadAll:
				c.reloadUnits()
			}
			// set global action to default
			unitManager.SetGlobalAction(globals.Act_Global_NoAction)
			// delay before next check
			time.Sleep(time.Duration(checkDelay) * time.Second)
			continue
		}

		// check all processes
		for _, proc := range c.processes {
			// update unit state
			proc.UpdateStatus()
			unit, err := unitManager.GetUnit(proc.Id)
			if err != nil {
				slog.Error("Failed to get unit while updating status", "id", proc.Id, "error", err.Error())
				continue
			}

			// check action for process
			if unit.Action != globals.Act_NoAction {
				var err error
				var actionStr string
				switch unit.Action {
				case globals.Act_Start:
					err = proc.Start()
					actionStr = "start"
				case globals.Act_Restart:
					err = proc.Restart()
					actionStr = "restart"
				case globals.Act_Stop:
					err = proc.Stop()
					actionStr = "stop"
				}
				if err != nil {
					slog.Error("Failed while applying action",
						"name", unit.Settings.Name, "act", actionStr, "error", err.Error())
				} else {
					slog.Info("Apply action for process", "name", unit.Settings.Name, "act", actionStr)
				}
				// set default action
				unitManager.SetUnitAction(proc.Id, globals.Act_NoAction)

				continue
			}

			// check process status
			switch unit.State {
			case globals.State_Failed:
				// if it failed update time
				unitManager.SetUnitStopTime(proc.Id, time.Now())
				// set waiting status
				unitManager.UpdateUnitState(proc.Id, globals.State_Timeout)
				slog.Warn("Process is failed, begin waiting for restart")
			case globals.State_Timeout:
				if unit.Settings.UseRestart {
					// if  process with restart check time of fail
					if time.Since(unit.StopTime) >= time.Duration(unit.Settings.RestartDelay) {
						slog.Info("Restarting process after fail", "name", unit.Settings.Name)
						err := proc.Restart()
						if err != nil {
							slog.Error("Failed to restart process", "name", unit.Settings.Name)
							unitManager.UpdateUnitState(proc.Id, globals.State_Failed)
						} else {
							slog.Info("Process was restarted after fail", "name", unit.Settings.Name)
						}
					}
				} else {
					unitManager.UpdateUnitState(proc.Id, globals.State_Stop)
					slog.Warn("Process in timeout state, but restart is not configured", "name", unit.Settings.Name)
				}
			}
		}
		// delay before next check
		time.Sleep(time.Duration(checkDelay) * time.Second)
	}

	slog.Info("Stop controller loop")

	// stop all processes
	c.stopAllUnit()

	// delete cache file
	cache.DeleteCacheFile()
}
