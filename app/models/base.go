package models

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"itodo/config"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var err error
var Db *sql.DB

func InitDataBase() {

	rootCertPool := x509.NewCertPool()
	pem, _ := ioutil.ReadFile("DigiCertGlobalRootCA.crt.pem")
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		log.Fatal("Failed to append PEM.")
	}
	mysql.RegisterTLSConfig("custom", &tls.Config{RootCAs: rootCertPool})

	connectionString := ""
	if os.Getenv("ITODOENV") != "local" {
		connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&allowNativePasswords=true&tls=custom", config.Config.MysqlUser, config.Config.MysqlPassword, config.Config.DataBaseHost, config.Config.DataBase)
	} else {
		connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&allowNativePasswords=true", config.Config.MysqlUser, config.Config.MysqlPassword, config.Config.DataBaseHost, config.Config.DataBase)
	}
	fmt.Println("aaaa")
	fmt.Println(connectionString)
	Db, err = sql.Open(config.Config.Driver, connectionString)

	if err != nil {
		log.Fatalln(err)
	}

	cmdU := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		id int(11) PRIMARY KEY AUTO_INCREMENT,
		uuid VARCHAR(36) NOT NULL UNIQUE,
		oauthid VARCHAR(36),
		vender int(11),
		created_at DATETIME)`, "users")
	Db.Exec(cmdU)

	cmdS := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		id int(11) PRIMARY KEY AUTO_INCREMENT,
		uuid VARCHAR(36) NOT NULL UNIQUE,
		user_id INT(11),
		created_at DATETIME)`, "sessions")
	Db.Exec(cmdS)

	cmdT := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		id int(11) PRIMARY KEY AUTO_INCREMENT,
		content VARCHAR(256) ,
		user_id INT(11),
		created_at DATETIME)`, "todos")
	Db.Exec(cmdT)

	cmdI := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		id int(11) PRIMARY KEY AUTO_INCREMENT,
		uuid VARCHAR(36) NOT NULL UNIQUE,
		content VARCHAR(256),
		todo_id INT(11),
		priority INT(1) DEFAULT 0,
		created_at DATETIME)`, "items")
	Db.Exec(cmdI)

	cmdV := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		id int(11) PRIMARY KEY AUTO_INCREMENT,
		uuid VARCHAR(36) NOT NULL UNIQUE,
		content JSON DEFAULT NULL,
		todo_id INT(11),
		created_at DATETIME)`, "elements")
	Db.Exec(cmdV)
}

func createUUID() (uuidobj uuid.UUID) {
	uuidobj, _ = uuid.NewUUID()
	return uuidobj
}

func Encrypt(plaintext string) (cryptext string) {
	cryptext = fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
	return cryptext
}
