package models

import (
	"log"
	"time"
)

type Item struct {
	ID        string
	Content   string
	TodoID    int
	Priority  int
	CreatedAt time.Time
}

func CreateItem(content string, todoId int) (err error) {
	cmd := `insert into items (
		uuid,
		content, 
		todo_id, 
		priority, 
		created_at) values (?, ?, ?, 0, ?)`

	_, err = Db.Exec(cmd, createUUID(), content, todoId, time.Now())
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func GetItems(todoId int) (items []Item, err error) {
	cmd := `select id, content, todo_id, priority, created_at from items where todo_id = ? order by priority desc, content`
	rows, err := Db.Query(cmd, todoId)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var item Item
		err = rows.Scan(&item.ID,
			&item.Content,
			&item.TodoID,
			&item.Priority,
			&item.CreatedAt)
		if err != nil {
			log.Fatalln(err)
		}
		items = append(items, item)
	}
	rows.Close()
	return items, err
}

func UpdateItemPriority(priority int, itemId int, todoId int) (err error) {
	cmd := `update items set priority = ? where id = ? and todo_id = ?`
	_, err = Db.Query(cmd, priority, itemId, todoId)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func DeleteItem(itemId int, todoId int) (err error) {
	cmd := `delete from items where id = ? and todo_id = ?`
	_, err = Db.Query(cmd, itemId, todoId)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}
