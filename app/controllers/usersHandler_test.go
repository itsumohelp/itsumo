package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type userModelStub struct {
	uuid string
}

func (u *userModelStub) Fetch(id int)                                {}
func (u *userModelStub) GetName() string                             { return u.uuid }
func (u *userModelStub) GetUuId() string                             { return u.uuid }
func (u *userModelStub) CreateUser() error                           { return nil }
func (u *userModelStub) GetUserByOAUTHID(oauthid string, vender int) {}
func (u *userModelStub) GetUserId() int                              { return 1 }

func Test_Getreq_Ok(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/req", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	u := &userModelStub{uuid: "test"}
	h := IniHandler(u)
	if assert.NoError(t, h.GetReq(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	}
}

func Test_Getreq_404(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/req", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	u := &userModelStub{uuid: ""}
	h := IniHandler(u)
	err := h.GetReq(c)

	if assert.NotNil(t, err) {
		err, res := err.(*echo.HTTPError)
		if res {
			assert.Equal(t, http.StatusNotFound, err.Code)
			assert.Equal(t, "Not Found", err.Message)
		}
	}
}
