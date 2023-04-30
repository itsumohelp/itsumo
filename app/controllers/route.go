package controllers

import (
	"itodo/app/models"
	"itodo/config"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

func Requestroute() {
	router := NewRouter()
	router.Logger.Fatal(router.Start(":" + config.Port))
}

func NewRouter() *echo.Echo {
	e := echo.New()
	e.Static("/static", "static")
	e.File("/", "static/top.html")
	e.GET("/get", func(ctx echo.Context) error {
		db := models.InitDataBase()
		u := models.NewUserModel(db)
		h := IniHandler(u)
		return h.GetReq(ctx)
	})
	e.GET("/socialauth", getOpenIdConnect)
	e.GET("/auth/google/callback/", getGoogleCallback)
	e.GET("/todo", func(ctx echo.Context) error {
		sessionHandler := IniSessionHandler(models.NewSessionModel(models.InitDataBase()))
		if !sessionHandler.checklogin(ctx) {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		return ctx.File("static/todos.html")
	})
	e.GET("/todos", func(ctx echo.Context) error {
		sessionHandler := IniSessionHandler(models.NewSessionModel(models.InitDataBase()))
		if !sessionHandler.checklogin(ctx) {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		todoHandler := IniTodoHandler(models.NewTodoModel(models.InitDataBase()))
		return todoHandler.getTodos(ctx, sessionHandler.session.GetSession().Userid)

	})
	e.POST("/todos/add", func(ctx echo.Context) error {
		sessionHandler := IniSessionHandler(models.NewSessionModel(models.InitDataBase()))
		if !sessionHandler.checklogin(ctx) {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		todoHandler := IniTodoHandler(models.NewTodoModel(models.InitDataBase()))
		return todoHandler.createTodo(ctx, sessionHandler.session.GetSession().Userid)
	})

	e.GET("/todos/:todoid", func(ctx echo.Context) error {
		sessionHandler := IniSessionHandler(models.NewSessionModel(models.InitDataBase()))
		if !sessionHandler.checklogin(ctx) {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		return ctx.File("static/item.html")
	})
	e.GET("/elements/:todoid", func(ctx echo.Context) error {
		sessionHandler := IniSessionHandler(models.NewSessionModel(models.InitDataBase()))
		if !sessionHandler.checklogin(ctx) {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		elementHandler := IniElementHandler(models.NewElementModel(models.InitDataBase()))
		return elementHandler.getElement(ctx)

	})
	e.POST("/elements/:todoid", func(ctx echo.Context) error {
		sessionHandler := IniSessionHandler(models.NewSessionModel(models.InitDataBase()))
		if !sessionHandler.checklogin(ctx) {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		elementHandler := IniElementHandler(models.NewElementModel(models.InitDataBase()))
		return elementHandler.createElement(ctx)

	})
	return e
}

// func getElements(c echo.Context) error {
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
// 		elements, _ := models.GetElements(todo.ID)
// 		return json.NewEncoder(c.Response()).Encode(elements)
// 	}
// }
// func postElements(c echo.Context) error {
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

// 		u := new([]models.Postele)
// 		if err := c.Bind(u); err != nil {
// 			return err
// 		}
// 		bytes, err := json.Marshal(u)
// 		models.UpdateElements(string(bytes), todo.ID)
// 		return json.NewEncoder(c.Response()).Encode(u)
// 	}
// }

// func signupData(user_id int) error {
// 	todo := new(models.Todo)
// 	todo.Content = "いつもヘルプへようこそ！"
// 	todo.UserID = user_id
// 	models.AddTodo(user_id, todo.Content)
// 	gettodo, _ := models.GetTodo(todo.UserID)

// 	element := new(models.Element)
// 	element.Content = "[{\"Value\":\"こんにちわ！\",\"Check\":0},{\"Value\":\"いつもをぜひ使ってください\",\"Check\":1}]"
// 	element.TodoID = gettodo.ID
// 	models.CreateElement(element.Content, element.TodoID)
// 	return nil
// }

func writeCookie(c echo.Context, name, value string) error {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)
	return nil
}
