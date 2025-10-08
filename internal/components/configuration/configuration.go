package configuration

import (
	"encoding/xml"
	"log/slog"
	"os"
	"path/filepath"
	"service_orchestrator/internal/components/globals"
	"service_orchestrator/internal/components/logger"
	"strings"
)

type configuration struct {
	XmlName  xml.Name `xml:"configuration"`
	Port     int      `xml:"http_port"`
	DebugLog bool     `xml:"debug_log"`
	Vars     []struct {
		ID    string `xml:"id,attr"`
		Value string `xml:",chardata"`
	} `xml:"vars>var"`
	GlobalSettings struct {
		AutoStart bool `xml:"auto_start"`
		Restart   struct {
			UseRestart bool `xml:",chardata"`
			Delay      uint `xml:"delay,attr"`
		} `xml:"restart"`
	} `xml:"global_settings"`
	Units []struct {
		Name    string `xml:"name,attr"`
		Cmd     string `xml:"cmd"`
		Args    string `xml:"args"`
		Restart struct {
			UseRestart *bool `xml:",chardata"`
			Delay      *uint `xml:"delay,attr"`
		} `xml:"restart"`
		AutoStart *bool `xml:"auto_start"`
	} `xml:"units>unit"`
}

func ReadConfiguration(logInstance *logger.Logger) (*[]globals.UnitSettings, error) {
	configPath := filepath.Join(globals.BinPath, "/orchestrator_assets/configuration.xml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Error while reading configuration.xml", "path", configPath, "error", err.Error())
		return nil, err
	}

	var config configuration
	err = xml.Unmarshal(data, &config)
	if err != nil {
		slog.Error("Error while parsing configuration.xml", "path", configPath, "error", err.Error())
		return nil, err
	}

	// set global settings
	// update logLevel
	if config.DebugLog {
		globals.LogLevel = logger.LogDebug
	} else {
		globals.LogLevel = logger.LogInfo
	}

	// set http port
	if globals.HttpPort == 0 {
		globals.HttpPort = config.Port
		slog.Info("Set http port", "port", config.Port)
	}

	// create map of the macros
	vars := make(map[string]string)
	for _, v := range config.Vars {
		vars[v.ID] = v.Value
	}

	// make unit list
	var units []globals.UnitSettings
	for _, unit := range config.Units {
		// replace macro
		cmd := replaceVars(unit.Cmd, vars)

		// split args
		args := strings.Fields(replaceVars(unit.Args, vars))

		// Check global settings
		autoStart := config.GlobalSettings.AutoStart
		if unit.AutoStart != nil {
			autoStart = *unit.AutoStart
		}

		useRestart := config.GlobalSettings.Restart.UseRestart
		if unit.Restart.UseRestart != nil {
			useRestart = *unit.Restart.UseRestart
		}

		restartDelay := config.GlobalSettings.Restart.Delay
		if unit.Restart.Delay != nil {
			restartDelay = *unit.Restart.Delay
		}

		// add new unit
		unitSettings := globals.UnitSettings{
			Name:         unit.Name,
			Cmd:          cmd,
			Args:         args,
			UseRestart:   useRestart,
			RestartDelay: restartDelay,
			AutoStart:    autoStart,
		}
		units = append(units, unitSettings)
	}
	return &units, nil
}

// replaceVars replace macro (example $src) to value
func replaceVars(input string, vars map[string]string) string {
	result := input
	for id, value := range vars {
		result = strings.ReplaceAll(result, "$"+id, value)
	}
	return result
}
