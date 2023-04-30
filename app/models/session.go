package models

import (
	"database/sql"
	"time"
)

type (
	Session interface {
		CreateSession(userid int)
		CheckSession(sessinUuid string) bool
		GetSession() SessionModel
	}
	SessionModel struct {
		db         *sql.DB
		Id         int
		Uuid       string
		Userid     int
		Created_at time.Time
	}
)

func NewSessionModel(db *sql.DB) *SessionModel {
	return &SessionModel{
		db: db,
	}
}

func (s *SessionModel) CreateSession(userid int) {
	var cmd string = `delete from sessions where user_id = ?`
	_, deleteErr := s.db.Exec(cmd, userid)
	if deleteErr != nil {
		panic(deleteErr.Error())
	}

	cmd1 := `insert into sessions (uuid, user_id, created_at) values (?, ?, ?)`
	_, insertErr := s.db.Exec(cmd1, createUUID(), userid, time.Now())
	if insertErr != nil {
		panic(insertErr.Error())
	}
	cmd2 := `select uuid from sessions where user_id = ?`
	err := s.db.QueryRow(cmd2, userid).Scan(
		&s.Uuid,
	)
	if err != nil {
		panic(err.Error())
	}
}

func (s *SessionModel) CheckSession(sessinUuid string) bool {
	cmd := `select id, uuid, user_id, created_at from sessions where uuid = ?`
	err := s.db.QueryRow(cmd, sessinUuid).Scan(
		&s.Id,
		&s.Uuid,
		&s.Userid,
		&s.Created_at)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		panic(err.Error())
	}
	return true
}
func (s *SessionModel) GetSession() SessionModel {
	return *s
}

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
