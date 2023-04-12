package models

import (
	"log"
	"time"
)

type User struct {
	ID        int
	UUID      string
	OAUTHID   string
	VENDER    int
	CreatedAt time.Time
	Todos     []Todo
}

type Session struct {
	ID        int
	UUID      string
	UserID    int
	CreatedAt time.Time
}

func (u *User) CreateUser() (err error) {
	cmd := `insert into users (
		uuid,
		oauthid,
		vender,
		created_at) values (?, ?, ?, ?);`

	_, err = Db.Exec(cmd,
		createUUID(),
		u.OAUTHID,
		u.VENDER,
		time.Now())

	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func GetUser(id int) (user User, err error) {
	user = User{}
	cmd := `select id, uuid, oauthid, vender, created_at
	from users where id = ?`
	err = Db.QueryRow(cmd, id).Scan(
		&user.ID,
		&user.UUID,
		&user.OAUTHID,
		&user.VENDER,
		&user.CreatedAt,
	)
	return user, err
}

func (u *User) CreateSession() (session Session, err error) {
	session = Session{}

	cmd1 := `insert into sessions (
		uuid, 
		user_id, 
		created_at) values (?, ?, ?)`

	_, err = Db.Exec(cmd1, createUUID(), u.ID, time.Now())
	if err != nil {
		log.Println(err)
	}

	cmd2 := `select id, uuid, user_id, created_at
	 from sessions where user_id = ?`

	err = Db.QueryRow(cmd2, u.ID).Scan(
		&session.ID,
		&session.UUID,
		&session.UserID,
		&session.CreatedAt)

	return session, err
}

func (sess *Session) CheckSession() (valid bool, err error) {
	cmd := `select id, uuid, user_id, created_at
	 from sessions where uuid = ?`

	err = Db.QueryRow(cmd, sess.UUID).Scan(
		&sess.ID,
		&sess.UUID,
		&sess.UserID,
		&sess.CreatedAt)

	if err != nil {
		valid = false
		return
	}
	if sess.ID != 0 {
		valid = true
	}
	return valid, err
}

func (sess *Session) DeleteSessionByUUID() (err error) {
	cmd := `delete from sessions where uuid = ?`
	_, err = Db.Exec(cmd, sess.UUID)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func (sess *Session) DeleteSessionByID() (err error) {
	cmd := `delete from sessions where user_id = ?`
	_, err = Db.Exec(cmd, sess.UserID)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func GetSessionByUserID(userId int) (sess Session, err error) {
	cmd := `select uuid, user_id, created_at from sessions where user_id = ?`
	err = Db.QueryRow(cmd, userId).Scan(
		&sess.UUID,
		&sess.UserID,
		&sess.CreatedAt)
	return sess, err
}

func (u *User) GetUserByOAUTHID(oauthid string) (user User, err error) {
	user = User{}
	cmd := `select id, uuid, oauthid, vender, created_at FROM users
	where oauthid = ?`
	err = Db.QueryRow(cmd, oauthid).Scan(
		&user.ID,
		&user.UUID,
		&user.OAUTHID,
		&user.VENDER,
		&user.CreatedAt)
	return user, err
}

func (sess *Session) GetUserBySession() (user User, err error) {
	user = User{}
	cmd := `select id, uuid, oauthid, created_at FROM users
	where id = ?`
	err = Db.QueryRow(cmd, sess.UserID).Scan(
		&user.ID,
		&user.UUID,
		&user.OAUTHID,
		&user.CreatedAt)
	return user, err
}
