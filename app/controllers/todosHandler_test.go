package controllers

import (
	"encoding/json"
	"fmt"
	"itodo/app/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type sessionModelStub struct {
	result bool
}

type todoModelStub struct {
	result error
}

func (u *sessionModelStub) CreateSession(int)        {}
func (u *sessionModelStub) CheckSession(string) bool { return u.result }
func (u *sessionModelStub) GetSession() models.SessionModel {
	return models.SessionModel{}
}

var todomodels models.TodoModel = models.TodoModel{Id: 1, Content: "test", UserId: "2"}

func (todo *todoModelStub) GetTodo() models.TodoModel    { return todomodels }
func (todo *todoModelStub) CreateTodo(string, int) error { return todo.result }
func (todo *todoModelStub) Fetch(int) []models.TodoModel { return []models.TodoModel{} }

func Test_checklogin_success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/todo", nil)
	req.Header.Add("Cookie", fmt.Sprintf("%s=%s", "chech", "test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	session := &sessionModelStub{result: true}
	h := IniSessionHandler(session)
	bool := h.checklogin(c)
	assert.True(t, bool)
}

func Test_checklogin_failer(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/todo", nil)
	req.Header.Add("Cookie", fmt.Sprintf("%s=%s", "chech", "test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	session := &sessionModelStub{result: false}
	h := IniSessionHandler(session)
	bool := h.checklogin(c)
	assert.False(t, bool)
}

func Test_createTodo(t *testing.T) {
	var request models.TodoReq = models.TodoReq{Content: "test"}
	jsonString, _ := json.Marshal(request)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/todo/add", strings.NewReader(string(jsonString)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Add("Cookie", fmt.Sprintf("%s=%s", "chech", "test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	todo := &todoModelStub{result: nil}
	h := IniTodoHandler(todo)
	err := h.createTodo(c, 1)
	if assert.NoError(t, err) {
		e, _ := json.Marshal(todomodels)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, string(e), rec.Body.String())
	}
}
