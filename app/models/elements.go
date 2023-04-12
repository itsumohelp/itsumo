package models

import (
	"log"
	"time"
)

type Postele struct {
	Value string `json:"value`
	Check int    `json:"check`
}

type Element struct {
	ID        string
	UUID      string
	Content   string
	TodoID    int
	Priority  int
	CreatedAt time.Time
}

func CreateElement(content string, todoId int) (err error) {
	cmd := `insert into elements (
		uuid,
		content, 
		todo_id, 
		created_at) values (?, ?, ?, ?)`

	_, err = Db.Exec(cmd, createUUID(), content, todoId, time.Now())
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func GetElements(todoId int) (elements []Element, err error) {
	cmd := `select id, content, todo_id, created_at from elements where todo_id = ?`
	rows, err := Db.Query(cmd, todoId)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var element Element
		err = rows.Scan(&element.ID,
			&element.Content,
			&element.TodoID,
			&element.CreatedAt)
		if err != nil {
			log.Fatalln(err)
		}
		elements = append(elements, element)
	}
	rows.Close()
	return elements, err
}

func UpdateElements(content string, todo_id int) (err error) {
	cmd := `update elements set content = ? where todo_id = ?`
	_, err = Db.Query(cmd, content, todo_id)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func DeleteElements(itemId int, todoId int) (err error) {
	cmd := `delete from items where id = ? and todo_id = ?`
	_, err = Db.Query(cmd, itemId, todoId)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}
