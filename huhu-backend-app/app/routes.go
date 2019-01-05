package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/heroku/huhu-backend-app/app/controllers"
	"github.com/heroku/huhu-backend-app/config"
	"github.com/heroku/huhu-backend-app/migrate"
	"github.com/jinzhu/gorm"
)

type App struct {
	Router *mux.Router
	DB     *gorm.DB
}

func (a *App) Initialize(config *config.Config) {
	dbURI := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=%s&parseTime=true",
		config.DB.Username,
		config.DB.Password,
		config.DB.Hostname,
		config.DB.Name,
		config.DB.Charset)

	db, err := gorm.Open(config.DB.Dialect, dbURI)
	if err != nil {
		log.Fatal("Could not connect database: " + dbURI)
	}

	a.DB = migrate.DBMigrate(db)
	a.Router = mux.NewRouter()
	a.setRouters()
}

func (a *App) setRouters() {
	a.Post("/login", a.Login)
	a.Post("/users/verify_email", a.VerifyByEmail)
	a.Post("/users/verify_reset_key", a.VerifyByResetKey)
	a.Post("/users", a.InputUser)
	a.Get("/users", a.ListUser)
	a.Get("/users/{id:[1-9]+}", a.OneUser)
	a.Put("/users/{id:[1-9]+}", a.UpdateUser)
	a.Delete("/users/{id:[1-9]+}", a.DeletedUser)
	a.Get("/", a.Index)
}

func (a *App) Post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("POST")
}

func (a *App) Get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("Get")
}

func (a *App) Put(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("Put")
}

func (a *App) Delete(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("Delete")
}

func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	controllers.Index(a.DB, w, r)
}

func (a *App) InputUser(w http.ResponseWriter, r *http.Request) {
	controllers.InputUser(a.DB, w, r)
}

func (a *App) ListUser(w http.ResponseWriter, r *http.Request) {
	controllers.ListUser(a.DB, w, r)
}

func (a *App) OneUser(w http.ResponseWriter, r *http.Request) {
	controllers.OneUser(a.DB, w, r)
}

func (a *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	controllers.UpdateUser(a.DB, w, r)
}

func (a *App) DeletedUser(w http.ResponseWriter, r *http.Request) {
	controllers.DeletedUser(a.DB, w, r)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	controllers.Login(a.DB, w, r)
}

func (a *App) VerifyByEmail(w http.ResponseWriter, r *http.Request) {
	controllers.VerifyByEmail(a.DB, w, r)
}

func (a *App) VerifyByResetKey(w http.ResponseWriter, r *http.Request) {
	controllers.VerifyByResetKey(a.DB, w, r)
}

func (a *App) Run(host string) {
	log.Fatal(http.ListenAndServe(host, a.Router))
}
