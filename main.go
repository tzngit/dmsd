package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/codeskyblue/go-sh"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	utils "github.com/tzngit/goutils"
)

type Config struct {
	Ip   string
	Port string
}

type DMSD struct {
	config    *Config
	curAbsDir string
}

func NewDMSD() *DMSD {
	return &DMSD{}
}

func (app *DMSD) LoadConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.SetDefault("application.ip", utils.LocalIp())
	c := Config{}
	err = viper.Unmarshal(&c)
	if err != nil {
		return err
	}
	app.config = &c
	app.curAbsDir = utils.CurAbsDir()
	return nil
}

func (app *DMSD) SetRoute(r *mux.Router) {
	fs := http.FileServer(http.Dir("."))
	r.PathPrefix("/static").Handler(fs)

	r.HandleFunc("/", app.IndexPage).Methods("GET")
	r.HandleFunc("/parse", app.ParseURL).Methods("POST")
}

func (app *DMSD) IndexPage(w http.ResponseWriter, r *http.Request) {
	pageContent := utils.File2String(filepath.Join(app.curAbsDir, "static", "html", "index.html"))
	fmt.Fprintf(w, pageContent)
}

type PostParam struct {
	URL string
}

type SongRecord struct {
	Name   string
	RawUrl string
	Cover  string
	Id     string
}

func (app *DMSD) ParseURL(w http.ResponseWriter, r *http.Request) {
	pp := PostParam{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&pp)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	err = sh.Command("wget", "-q", pp.URL, "-O", "/tmp/t.html").Run()
	if err != nil {
		log.Println("wget fail", pp.URL, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var output []byte
	output, err = sh.Command("grep", "song_records = ", "/tmp/t.html").Command("awk", "-F", "records = ", "{print $2}").Command("sed", "s/,$//g").Output()
	if err != nil {
		log.Println("exec cat awk sed error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var song_records []SongRecord
	if err = json.Unmarshal(output, &song_records); err != nil {
		log.Println("json unmarshal fail,", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	utils.ResponseJson(w, song_records)
}

func main() {
	app := NewDMSD()
	r := mux.NewRouter()
	err := app.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.SetOutput(os.Stdout)
	app.SetRoute(r)

	cfg := app.config
	fmt.Println("Douban Music Site Downlaoder start!")
	err = http.ListenAndServe(cfg.Ip+":"+cfg.Port, r)
	if err != nil {
		panic(err)
	}
}
