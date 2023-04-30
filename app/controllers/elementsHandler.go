package controllers

import (
	"encoding/json"
	"itodo/app/models"
	"strconv"

	"github.com/labstack/echo"
)

type (
	elementHandler struct {
		element models.Element
	}
)

func IniElementHandler(e models.Element) *elementHandler {
	return &elementHandler{e}
}

func (eh elementHandler) getElement(c echo.Context) error {
	todoid, _ := strconv.Atoi(c.Param("todoid"))
	eh.element.Fetch(todoid)
	return json.NewEncoder(c.Response()).Encode(eh.element.GetElement())
}
func (eh elementHandler) createElement(c echo.Context) error {
	todoid, _ := strconv.Atoi(c.Param("todoid"))
	var elementReq *[]models.ElementReq = new([]models.ElementReq)
	if err := c.Bind(elementReq); err != nil {
		return echo.ErrInternalServerError
	}
	bytes, _ := json.Marshal(elementReq)

	var registed bool = eh.element.CheckElement(todoid)
	if registed {
		eh.element.UpdateElement(string(bytes), todoid)
	} else {
		eh.element.CreateElement(string(bytes), todoid)
	}

	return json.NewEncoder(c.Response()).Encode("")
}
