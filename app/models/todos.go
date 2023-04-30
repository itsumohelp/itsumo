package models

import (
	"database/sql"
	"time"
)

type (
	Todo interface {
		Fetch(int) []TodoModel
		GetTodo() TodoModel
		CreateTodo(string, int) error
	}
	TodoModel struct {
		db         *sql.DB
		Id         int
		Content    string
		UserId     string
		Created_at time.Time
	}
	TodoReq struct {
		Content string `json:"content"`
	}
)

func NewTodoModel(db *sql.DB) *TodoModel {
	return &TodoModel{
		db: db,
	}
}

func (todo *TodoModel) CreateTodo(content string, userid int) error {
	cmd := `insert into todos (
		content,
		user_id,
		created_at) values (?, ?, ?)`
	res, err := todo.db.Exec(cmd, content, userid, time.Now())
	inertid, _ := res.LastInsertId()
	todo.Id = int(inertid)
	todo.Content = content
	return err
}

// func AddTodo(user_id int, content string) (err error) {
// 	cmd := `insert into todos (
// 		content,
// 		user_id,
// 		created_at) values (?, ?, ?)`

// 	_, err = Db.Exec(cmd, content, user_id, time.Now())
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return err
// }

func (todo *TodoModel) Fetch(user_id int) []TodoModel {
	cmd := `select id, content, user_id, created_at from todos where user_id = ? order by content`
	rows, err := todo.db.Query(cmd, user_id)

	switch {
	case err == sql.ErrNoRows:
		return []TodoModel{}
	case err != nil:
		panic(err.Error())
	}

	var todos []TodoModel
	for rows.Next() {
		var todo TodoModel
		err = rows.Scan(
			&todo.Id,
			&todo.Content,
			&todo.UserId,
			&todo.Created_at)
		if err != nil {
			panic(err.Error())
		}
		todos = append(todos, todo)
	}
	return todos
}

func (todo *TodoModel) GetTodo() TodoModel {
	return *todo
}

// func GetTodos(id int, user_id int) (todo Todo, err error) {
// 	cmd := `select id, content, user_id, created_at from todos
// 	where id = ? and user_id = ? order by content`
// 	todo = Todo{}

// 	err = Db.QueryRow(cmd, id, user_id).Scan(
// 		&todo.ID,
// 		&todo.Content,
// 		&todo.UserID,
// 		&todo.CreatedAt)

// 	return todo, err
// }

// func GetAllTodos() (todos []Todo, err error) {
// 	cmd := `select id, content, user_id, created_at from todos order by content`
// 	rows, err := Db.Query(cmd)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	for rows.Next() {
// 		var todo Todo
// 		err = rows.Scan(&todo.ID,
// 			&todo.Content,
// 			&todo.UserID,
// 			&todo.CreatedAt)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		todos = append(todos, todo)
// 	}
// 	rows.Close()

// 	return todos, err
// }

// func (u *User) GetTodosByUser() (todos []Todo, err error) {
// 	cmd := `select id, content, user_id, created_at from todos
// 	where user_id = ? order by content`

// 	rows, err := Db.Query(cmd, u.ID)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	for rows.Next() {
// 		var todo Todo
// 		err = rows.Scan(
// 			&todo.ID,
// 			&todo.Content,
// 			&todo.UserID,
// 			&todo.CreatedAt)

// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		todos = append(todos, todo)
// 	}
// 	rows.Close()

// 	return todos, err
// }

// func (t *Todo) UpdateTodo() error {
// 	cmd := `update todos set content = ?, user_id = ?
// 	where id = ?`
// 	_, err = Db.Exec(cmd, t.Content, t.UserID, t.ID)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return err
// }

// func DeleteTodo(todoid int) error {
// 	cmdI := `delete from items where todo_id = ?`
// 	_, err = Db.Exec(cmdI, todoid)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	cmdT := `delete from todos where id = ?`
// 	_, err = Db.Exec(cmdT, todoid)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return err
// }
