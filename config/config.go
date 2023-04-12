package config

import (
	"context"
	"itodo/utils"
	"log"
	"os"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"gopkg.in/go-ini/ini.v1"
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

var Config ConfigList
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

	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln(err)
	}
	Config = ConfigList{
		Port:          cfg.Section("web").Key("port").MustString("8080"),
		Driver:        cfg.Section("db").Key("driver").String(),
		DataBaseHost:  DataBaseHost,
		DataBase:      DataBase,
		MysqlUser:     MysqlUser,
		MysqlPassword: MysqlPassword,
		LogFile:       cfg.Section("web").Key("logfile").String(),
		Static:        cfg.Section("web").Key("static").String(),
	}
	Authconfig = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  Host + "/auth/google/callback/",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	utils.LoggingSettings(cfg.Section("web").Key("logfile").String())
}
