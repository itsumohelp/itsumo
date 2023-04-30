package controllers

import (
	"encoding/json"
	"itodo/app/models"

	"github.com/labstack/echo"
)

type (
	todoHandler struct {
		todo models.Todo
	}
)

func IniTodoHandler(u models.Todo) *todoHandler {
	return &todoHandler{u}
}

func (s sessionHandler) checklogin(c echo.Context) bool {
	cookie, _ := c.Cookie("chech")
	return s.session.CheckSession(cookie.Value)
}

func (th todoHandler) getTodos(c echo.Context, userid int) error {
	var result []models.TodoModel = th.todo.Fetch(userid)
	return json.NewEncoder(c.Response()).Encode(result)
}

func (th todoHandler) createTodo(c echo.Context, userid int) error {
	var todoReq *models.TodoReq = new(models.TodoReq)
	if err := c.Bind(todoReq); err != nil {
		return echo.ErrInternalServerError
	}
	if err := th.todo.CreateTodo(todoReq.Content, userid); err != nil {
		return echo.ErrInternalServerError
	}
	return json.NewEncoder(c.Response()).Encode(th.todo.GetTodo())
}

// func getTodoList(c echo.Context) error {
// 	cookie, _ := c.Cookie("chech")
// 	sess, err := Checksession(cookie.Value)
// 	if err != nil {
// 		return c.Redirect(http.StatusFound, "/")
// 	} else {
// 		todoid, _ := strconv.Atoi(c.Param("todoid"))
// 		user, err := sess.GetUserBySession()
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		todo, _ := models.GetTodos(todoid, user.ID)
// 		items, _ := models.GetItems(todo.ID)
// 		return json.NewEncoder(c.Response()).Encode(items)
// 	}
// }
