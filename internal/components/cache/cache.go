package cache

import (
	"encoding/xml"
	"log/slog"
	"os"
	"service_orchestrator/internal/components/globals"
	"strings"
	"sync"
)

type CacheUnit struct {
	XmlName xml.Name `xml:"units"`
	Units   []struct {
		Pid  int    `xml:"pid,attr"`
		Name string `xml:"name,attr"`
		Cmd  string `xml:"cmd"`
		Args string `xml:"args"`
	} `xml:"unit"`
}

var cacheMtx sync.Mutex

var cachePath string = "/tmp/orchestrator/cache.xml"

func createFile() bool {
	_, err := os.Stat(cachePath)
	if os.IsNotExist(err) {
		_, err = os.Create(cachePath)
		if err != nil {
			slog.Error("Failed to create cache file", "error", err.Error())
			return false
		}
	}
	return true
}

func LoadUnits() CacheUnit {
	cacheMtx.Lock()
	defer cacheMtx.Unlock()

	if !createFile() {
		return CacheUnit{}
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		slog.Error("Error while reading cache", "error", err.Error())
		return CacheUnit{}
	}

	var units CacheUnit
	err = xml.Unmarshal(data, &units)
	if err != nil {
		slog.Error("Error while parsing cache", "error", err.Error())
	}
	return units
}

func DeleteCacheFile() {
	cacheMtx.Lock()
	defer cacheMtx.Unlock()

	err := os.Remove(cachePath)
	if err != nil {
		slog.Error("Error while deleting cache file", "error", err.Error())
	}
}

func UpdateUnitInFile(unit globals.UnitSettings, pid int) {
	cacheMtx.Lock()
	defer cacheMtx.Unlock()

	if !createFile() {
		return
	}

	// Load existing units
	units := LoadUnits()

	// find unit by name
	updated := false
	for i := range units.Units {
		if units.Units[i].Name == unit.Name {
			units.Units[i].Pid = pid
			units.Units[i].Cmd = unit.Cmd
			units.Units[i].Args = strings.Join(unit.Args, " ")
			updated = true
			break
		}
	}

	if !updated {
		slog.Warn("Unit with name not found for updating", "pid", pid)
		return
	}

	// Marshal back to XML
	data, err := xml.MarshalIndent(units, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal updated units", "error", err.Error())
		return
	}

	// Write back to file
	err = os.WriteFile(cachePath, data, 0644)
	if err != nil {
		slog.Error("Failed to write updated cache file", "error", err.Error())
	}
}

func DeleteUnitFromFile(name string) {
	cacheMtx.Lock()
	defer cacheMtx.Unlock()

	if !createFile() {
		return
	}

	// Load existing units
	units := LoadUnits()
	found := false

	newUnits := make([]struct {
		Pid  int    `xml:"pid,attr"`
		Name string `xml:"name,attr"`
		Cmd  string `xml:"cmd"`
		Args string `xml:"args"`
	}, 0, len(units.Units))

	for _, unit := range units.Units {
		if unit.Name == name {
			found = true
			continue
		}
		newUnits = append(newUnits, unit)
	}

	if !found {
		slog.Warn("Unit with name not found for deleting from cache", "name", name)
		return
	}

	units.Units = newUnits

	data, err := xml.MarshalIndent(units, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal updated units", "error", err.Error())
		return
	}

	// Write back to file
	err = os.WriteFile(cachePath, data, 0644)
	if err != nil {
		slog.Error("Failed to write updated cache file", "error", err.Error())
	}
}
