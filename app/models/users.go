package models

import (
	"database/sql"
	"fmt"
	"time"
)

type (
	User interface {
		Fetch(int)
		GetUserByOAUTHID(string, int)
		CreateUser() error
		GetUserId() int
		GetUuId() string
	}
	UserModel struct {
		db         *sql.DB
		Id         int
		Uuid       string
		Oauthid    string
		Vender     int
		Created_at time.Time
	}
)

func NewUserModel(db *sql.DB) *UserModel {
	return &UserModel{
		db:         db,
		Id:         0,
		Uuid:       "",
		Oauthid:    "",
		Vender:     0,
		Created_at: time.Now(),
	}
}

func (u *UserModel) CreateUser() error {
	cmd := `insert into users (
		uuid,
		oauthid,
		vender,
		created_at) values (?, ?, ?, ?);`
	_, err := u.db.Exec(cmd, createUUID(), u.Oauthid, u.Vender, time.Now())
	return err
}

func (u *UserModel) GetUserId() int {
	return u.Id
}
func (u *UserModel) GetUuId() string {
	return u.Uuid
}

// func GetUser(id int) (user User, err error) {
// 	user = User{}
// 	cmd := `select id, uuid, oauthid, vender, created_at
// 	from users where id = ?`
// 	err = Db.QueryRow(cmd, id).Scan(
// 		&user.ID,
// 		&user.UUID,
// 		&user.OAUTHID,
// 		&user.VENDER,
// 		&user.CreatedAt,
// 	)
// 	return user, err
// }

// func (u *User) CreateSession() (session Session, err error) {
// 	session = Session{}
// 	fmt.Println("bbb")
// 	cmd1 := `insert into sessions (
// 		uuid,
// 		user_id,
// 		created_at) values (?, ?, ?)`

// 	_, err = Db.Exec(cmd1, createUUID(), u.ID, time.Now())
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	cmd2 := `select id, uuid, user_id, created_at
// 	 from sessions where user_id = ?`

// 	err = Db.QueryRow(cmd2, u.ID).Scan(
// 		&session.ID,
// 		&session.UUID,
// 		&session.UserID,
// 		&session.CreatedAt)

// 	return session, err
// }

// func (sess *Session) CheckSession() (valid bool, err error) {
// 	cmd := `select id, uuid, user_id, created_at
// 	 from sessions where uuid = ?`

// 	err = Db.QueryRow(cmd, sess.UUID).Scan(
// 		&sess.ID,
// 		&sess.UUID,
// 		&sess.UserID,
// 		&sess.CreatedAt)

// 	if err != nil {
// 		valid = false
// 		return
// 	}
// 	if sess.ID != 0 {
// 		valid = true
// 	}
// 	return valid, err
// }

// func (sess *Session) DeleteSessionByUUID() (err error) {
// 	cmd := `delete from sessions where uuid = ?`
// 	_, err = Db.Exec(cmd, sess.UUID)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return err
// }

// func (sess *Session) DeleteSessionByID() (err error) {
// 	cmd := `delete from sessions where user_id = ?`
// 	_, err = Db.Exec(cmd, sess.UserID)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return err
// }

// func GetSessionByUserID(userId int) (sess Session, err error) {
// 	cmd := `select uuid, user_id, created_at from sessions where user_id = ?`
// 	err = Db.QueryRow(cmd, userId).Scan(
// 		&sess.UUID,
// 		&sess.UserID,
// 		&sess.CreatedAt)
// 	return sess, err
// }

func (u *UserModel) GetUserByOAUTHID(oauthid string, vender int) {
	u.Oauthid = oauthid
	u.Vender = vender
	cmd := `select id, uuid, created_at FROM users where oauthid = ? and vender = ?`
	err := u.db.QueryRow(cmd, oauthid, vender).Scan(
		&u.Id,
		&u.Uuid,
		&u.Created_at)
	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		panic(err.Error())
	}
}

// func (sess *Session) GetUserBySession() (user User, err error) {
// 	user = User{}
// 	cmd := `select id, uuid, oauthid, created_at FROM users
// 	where id = ?`
// 	err = Db.QueryRow(cmd, sess.UserID).Scan(
// 		&user.ID,
// 		&user.UUID,
// 		&user.OAUTHID,
// 		&user.CreatedAt)
// 	return user, err
// }

func (u *UserModel) Fetch(id int) {
	if id < 1 {
		return
	}

	err := u.db.QueryRow("SELECT uuid FROM users WHERE id=? LIMIT 1", id).Scan(&u.Uuid)
	switch {
	case err == sql.ErrNoRows:
		u.Uuid = ""
	case err != nil:
		panic(err.Error())
	}
	fmt.Println("koredesu")
}
