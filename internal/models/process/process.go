package process

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"service_orchestrator/internal/components/globals"
	"service_orchestrator/internal/models/manager"
	"strings"
	"syscall"
	"time"
)

// struct for process declaration and control
type Process struct {
	Id  int
	Pid int
}

// check process status is running
func (u *Process) isRunning() bool {
	// read status file
	procStatPath := fmt.Sprintf("/proc/%d/status", u.Id)
	data, err := os.ReadFile(procStatPath)
	if err != nil {
		slog.Error("Error read status", "pid", u.Pid, "error", err.Error())
		return false
	}

	// parse lines
	procState := false
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "State:") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				letter := string(parts[1][0])
				if _, ok := globals.ProcSystemStates[letter]; ok {
					procState = globals.ProcSystemStates[letter] == globals.State_Work
				} else {
					slog.Error("Unknown state in status file", "pid", u.Pid, "status", letter)
				}
			}
			break
		}
	}

	return procState
}

// start process
func (u *Process) Start() error {
	// get current settings for this unit
	unitManager := manager.GetInstance()
	unit := unitManager.GetUnit(u.Id)

	if u.Pid != 0 && u.isRunning() {
		slog.Warn("Unit is already running", "name", unit.Settings.Name, "pid", u.Pid)
		unitManager.UpdateUnitState(u.Id, globals.State_Failed)
		return nil
	}

	// check path
	_, err := os.Stat(unit.Settings.Cmd)
	if os.IsNotExist(err) {
		slog.Error("Path doesn't exist, can't start unit", "name", unit.Settings.Name, "cmd", unit.Settings.Cmd)
		unitManager.UpdateUnitState(u.Id, globals.State_Failed)
		return err
	} else if err != nil {
		slog.Error("Error checking unit cmd path", "name", unit.Settings.Name, "cmd", unit.Settings.Cmd)
		unitManager.UpdateUnitState(u.Id, globals.State_Failed)
		return err
	}

	// set command
	cmd := exec.Command(unit.Settings.Cmd, unit.Settings.Args...)

	// create new session for process
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// try to start process
	err = cmd.Start()
	if err != nil {
		unitManager.UpdateUnitState(u.Id, globals.State_Failed)
		return err
	}

	// store pid
	u.Pid = cmd.Process.Pid

	// free resources
	err = cmd.Process.Release()
	if err != nil {
		slog.Error("Error while 'release' unit", "name", unit.Settings.Name, "cmd", unit.Settings.Cmd)
		return err
	}

	// update unit state
	unitManager.UpdateUnitState(u.Id, globals.State_Work)

	slog.Info("Start process", "name", unit.Settings.Name, "pid", u.Pid)

	return nil
}

// stop process by pid
func (u *Process) Stop() error {
	// get current settings for this unit
	unitManager := manager.GetInstance()
	unit := unitManager.GetUnit(u.Id)

	if u.Pid == 0 || !u.isRunning() {
		slog.Warn("Unit isn't running", "name", unit.Settings.Name, "pid", u.Pid)
		unitManager.UpdateUnitState(u.Id, globals.State_Stop)
		u.Pid = 0
		return nil
	}

	// check process
	proc, err := os.FindProcess(u.Pid)
	if err != nil {
		slog.Warn("Can not find process", "name", unit.Settings.Name, "pid", u.Pid)
		unitManager.UpdateUnitState(u.Id, globals.State_Stop)
		return err
	}

	// send sigterm
	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		slog.Warn("Error sending SIGTERM to unit", "name", unit.Settings.Name, "pid", u.Pid)
		unitManager.UpdateUnitState(u.Id, globals.State_Failed)
		return err
	}

	// waiting for 30 seconds while unit stop
	for range 30 {
		if !u.isRunning() {
			// unit was stopped
			u.Pid = 0
			slog.Info("Process was stopped", "name", unit.Settings.Name)
			unitManager.UpdateUnitState(u.Id, globals.State_Stop)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	// if we here, send sigkill
	slog.Warn("Unit doesn't stop, send SIGKILL", "name", unit.Settings.Name)
	err = proc.Signal(syscall.SIGKILL)
	if err != nil {
		slog.Warn("Error sending SIGTERM to unit", "name", unit.Settings.Name, "pid", u.Pid)
		unitManager.UpdateUnitState(u.Id, globals.State_Failed)
		return err
	}

	if !u.isRunning() {
		// unit was stopped
		u.Pid = 0
		slog.Info("Process was stopped", "name", unit.Settings.Name)
		unitManager.UpdateUnitState(u.Id, globals.State_Stop)
	}
	return nil
}

// restart process
func (u *Process) Restart() error {
	err := u.Stop()
	if err != nil {
		return err
	}
	return u.Start()
}

// update process status, check is it working still
func (u *Process) UpdateStatus() {
	// get current settings for this unit
	unitManager := manager.GetInstance()
	unit := unitManager.GetUnit(u.Id)
	if unit.State == globals.State_Work {
		// check process state
		if u.Pid != 0 && !u.isRunning() {
			// it should to work, if it doesn't work, than it's failed
			unitManager.UpdateUnitState(u.Id, globals.State_Failed)
		}
	}
}
