package models

import (
	"database/sql"
	"time"
)

type (
	Element interface {
		Fetch(int)
		CheckElement(int) bool
		GetElement() ElementModel
		CreateElement(string, int)
		UpdateElement(string, int)
	}
	ElementModel struct {
		db        *sql.DB
		Id        string
		Uuid      string
		Content   string
		TodoId    int
		CreatedAt time.Time
	}
	ElementReq struct {
		Value    string `json:"value`
		Check    int    `json:"check`
		Deadline string `json:"deadline"`
	}
)

func NewElementModel(db *sql.DB) *ElementModel {
	return &ElementModel{
		db: db,
	}
}

func (e *ElementModel) Fetch(todoId int) {
	cmd := `select id, uuid, content, todo_id, created_at from elements where todo_id = ?`
	err := e.db.QueryRow(cmd, todoId).Scan(&e.Id, &e.Uuid, &e.Content, &e.TodoId, &e.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		panic(err.Error())
	}
}

func (e *ElementModel) GetElement() ElementModel {
	return *e
}

func (e *ElementModel) CheckElement(todoId int) bool {
	cmd := `select id from elements where todo_id = ?`
	err := e.db.QueryRow(cmd, todoId).Scan(&e.Id)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		panic(err.Error())
	}
	return true
}

func (e *ElementModel) CreateElement(content string, todoId int) {
	cmd := `insert into elements (
		uuid,
		content,
		todo_id,
		created_at) values (?, ?, ?, ?)`

	_, err := e.db.Exec(cmd, createUUID(), content, todoId, time.Now())
	if err != nil {
		panic(err.Error())
	}
}

func (e *ElementModel) UpdateElement(content string, todoId int) {
	cmd := `update elements set content = ? where todo_id = ?`
	_, err := e.db.Exec(cmd, content, todoId)
	if err != nil {
		panic(err.Error())
	}
}

// func DeleteElements(itemId int, todoId int) (err error) {
// 	cmd := `delete from items where id = ? and todo_id = ?`
// 	_, err = Db.Query(cmd, itemId, todoId)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return err
// }
