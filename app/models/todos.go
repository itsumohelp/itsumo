package models

import (
	"log"
	"time"
)

type Todo struct {
	ID        int
	Content   string
	UserID    int
	CreatedAt time.Time
	Items     []Item
}

func (u *User) CreateTodo(content string) (err error) {
	cmd := `insert into todos (
		content, 
		user_id, 
		created_at) values (?, ?, ?)`

	_, err = Db.Exec(cmd, content, u.ID, time.Now())
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func AddTodo(user_id int, content string) (err error) {
	cmd := `insert into todos (
		content, 
		user_id, 
		created_at) values (?, ?, ?)`

	_, err = Db.Exec(cmd, content, user_id, time.Now())
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func GetTodo(user_id int) (todo Todo, err error) {
	cmd := `select id, content, user_id, created_at from todos
	where user_id = ? order by content`
	err = Db.QueryRow(cmd, user_id).Scan(
		&todo.ID,
		&todo.Content,
		&todo.UserID,
		&todo.CreatedAt)
	return todo, err
}

func GetTodos(id int, user_id int) (todo Todo, err error) {
	cmd := `select id, content, user_id, created_at from todos
	where id = ? and user_id = ? order by content`
	todo = Todo{}

	err = Db.QueryRow(cmd, id, user_id).Scan(
		&todo.ID,
		&todo.Content,
		&todo.UserID,
		&todo.CreatedAt)

	return todo, err
}

func GetAllTodos() (todos []Todo, err error) {
	cmd := `select id, content, user_id, created_at from todos order by content`
	rows, err := Db.Query(cmd)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID,
			&todo.Content,
			&todo.UserID,
			&todo.CreatedAt)
		if err != nil {
			log.Fatalln(err)
		}
		todos = append(todos, todo)
	}
	rows.Close()

	return todos, err
}

func (u *User) GetTodosByUser() (todos []Todo, err error) {
	cmd := `select id, content, user_id, created_at from todos
	where user_id = ? order by content`

	rows, err := Db.Query(cmd, u.ID)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var todo Todo
		err = rows.Scan(
			&todo.ID,
			&todo.Content,
			&todo.UserID,
			&todo.CreatedAt)

		if err != nil {
			log.Fatalln(err)
		}
		todos = append(todos, todo)
	}
	rows.Close()

	return todos, err
}

func (t *Todo) UpdateTodo() error {
	cmd := `update todos set content = ?, user_id = ? 
	where id = ?`
	_, err = Db.Exec(cmd, t.Content, t.UserID, t.ID)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func DeleteTodo(todoid int) error {
	cmdI := `delete from items where todo_id = ?`
	_, err = Db.Exec(cmdI, todoid)
	if err != nil {
		log.Fatalln(err)
	}

	cmdT := `delete from todos where id = ?`
	_, err = Db.Exec(cmdT, todoid)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}
