package handlers

import (
	"service_orchestrator/internal/components/globals"
	"service_orchestrator/internal/models/manager"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetAllProcesses(c *fiber.Ctx) error {
	units := manager.GetInstance().GetAllUnits()
	unitWithId := make([]struct {
		Id   int
		Unit globals.Unit
	}, 0, len(units))
	for id, unit := range units {
		unitWithId = append(unitWithId, struct {
			Id   int
			Unit globals.Unit
		}{
			Id:   id,
			Unit: unit,
		})
	}
	c.Status(fiber.StatusOK).JSON(unitWithId)
	return nil
}

func GetProcessById(c *fiber.Ctx) error {
	id := c.Params("id")
	unitId, err := strconv.Atoi(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		return nil
	}

	unit, err := manager.GetInstance().GetUnit(unitId)
	if err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		return nil
	}

	c.Status(fiber.StatusOK).JSON(unit)
	return nil
}

func AllProcessesAction(c *fiber.Ctx) error {
	// get command from query
	action := c.Query("action")
	// handle action
	procAction := globals.Act_NoAction
	switch action {
	case "start":
		procAction = globals.Act_Global_StartAll
	case "stop":
		procAction = globals.Act_Global_StopAll
	case "reload":
		procAction = globals.Act_Global_ReloadAll
	default:
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid action"})
		return nil
	}
	// set global action
	go manager.GetInstance().SetGlobalAction(procAction)

	c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "action started"})

	return nil
}

func ProcessActionById(c *fiber.Ctx) error {
	id := c.Params("id")
	unitId, err := strconv.Atoi(id)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		return nil
	}

	// get command from query
	action := c.Query("action")

	// handle command
	procAction := globals.Act_NoAction
	switch action {
	case "start":
		procAction = globals.Act_Start
	case "stop":
		procAction = globals.Act_Stop
	case "restart":
		procAction = globals.Act_Restart
	default:
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid action"})
		return nil
	}

	// set action for unit
	manager.GetInstance().SetUnitAction(unitId, procAction)

	c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "action started"})

	return nil
}
