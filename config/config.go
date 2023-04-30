package config

import (
	"context"
	"log"
	"os"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type ConfigList struct {
	Port          string
	Driver        string
	DataBaseHost  string
	DataBase      string
	MysqlUser     string
	MysqlPassword string
	LogFile       string
	Static        string
}

type Constant struct {
	GoogleOpenIDConnect   int
	CookieName            string
	GoogleAuthStateCookie string
	GoogleAuthNonceCookie string
}

var Config ConfigList
var Constants Constant
var Context context.Context
var Authconfig oauth2.Config
var Verifier *oidc.IDTokenVerifier
var clientID = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
var clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
var Port string = "80"

func LoadConfig() {
	var DataBaseHost string = os.Getenv("MYSQL_DATABASE_HOST")
	var DataBase string = os.Getenv("MYSQL_DATABASE")
	var MysqlUser string = os.Getenv("MYSQL_USER")
	var MysqlPassword string = os.Getenv("MYSQL_PASSWORD")
	var Host string = "https://itsumo.help"
	if os.Getenv("ITODOENV") == "local" {
		DataBaseHost = "localhost"
		DataBase = "itodo"
		MysqlUser = "itodoap"
		MysqlPassword = "Dbadmin2"
		Port = "5556"
		Host = "http://localhost:" + Port
	} else if os.Getenv("ITODOENV") == "dev" {
	}

	Context = context.Background()
	provider, err := oidc.NewProvider(Context, "https://accounts.google.com")

	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	Verifier = provider.Verifier(oidcConfig)

	Config = ConfigList{
		DataBaseHost:  DataBaseHost,
		DataBase:      DataBase,
		MysqlUser:     MysqlUser,
		MysqlPassword: MysqlPassword,
		Driver:        "mysql",
	}
	Authconfig = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  Host + "/auth/google/callback/",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	Constants = Constant{
		GoogleOpenIDConnect:   1,
		CookieName:            "chech",
		GoogleAuthStateCookie: "state",
		GoogleAuthNonceCookie: "nonce",
	}
}
