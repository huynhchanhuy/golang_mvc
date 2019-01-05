package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"crypto/tls"
	"encoding/gob"
	"github.com/gorilla/mux"
	// "github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/twinj/uuid"
	"gopkg.in/gomail.v2"

	"flag"
	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/google"
	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"

	_ "github.com/joho/godotenv/autoload"
)

var (
	APP_NAME                = os.Getenv("APP_NAME")
	WEB_PORT                = os.Getenv("PORT")
	API_URL                 = os.Getenv("API_URL")
	COOKIE_NAME             = os.Getenv("COOKIE_NAME")
	SESSION_USER_KEY        = os.Getenv("SESSION_USER_KEY")
	GOOGLE_SESSION_USER_KEY = os.Getenv("GOOGLE_SESSION_USER_KEY")
	GOOGLE_CALLBACK_URI     = os.Getenv("GOOGLE_CALLBACK_URI")
	GOOGLE_CALLBACK_API     = os.Getenv("GOOGLE_CALLBACK_HOST") + GOOGLE_CALLBACK_URI
	GOOGLE_CLIENT_ID        = os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_CLIENT_SECRET    = os.Getenv("GOOGLE_CLIENT_SECRET")
)

type Config struct {
	ClientID     string
	ClientSecret string
}

type User struct {
	ID            int
	Email         string
	Fullname      string
	Password      string
	Address       string
	Telephone     string
	ResetKey      string
	Authenticated bool
}

type Meta struct {
	message  string
	severity string
}

type UserResp struct {
	Status int
	Meta   Meta
	Data   User
}

func GetApiUrl(uri string) string {
	server := API_URL
	return server + uri
}

// store will hold all session data
var store *sessions.CookieStore

// tpl holds all parsed templates
var tpl *template.Template

func init() {
	// authKeyOne := securecookie.GenerateRandomKey(64)
	// encryptionKeyOne := securecookie.GenerateRandomKey(32)

	store = sessions.NewCookieStore(
		// authKeyOne,
		// encryptionKeyOne,
		[]byte(APP_NAME),
	)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 5,
		HttpOnly: true,
	}

	gob.Register(User{})

	tpl = template.Must(template.ParseGlob("view/*.gohtml"))
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", index)
	router.HandleFunc("/login", login)
	router.HandleFunc("/logout", logout)
	router.HandleFunc("/forbidden", forbidden)
	router.HandleFunc("/profile", profile)
	router.HandleFunc("/register", register)
	router.HandleFunc("/notice", notice)
	router.HandleFunc("/forgot", forgot)
	router.HandleFunc("/verify/{uuid}", verify)

	config := &Config{
		ClientID:     GOOGLE_CLIENT_ID,
		ClientSecret: GOOGLE_CLIENT_SECRET,
	}

	// allow consumer credential flags to override config fields
	clientID := flag.String("client-id", "", GOOGLE_CLIENT_ID)
	clientSecret := flag.String("client-secret", "", GOOGLE_CLIENT_SECRET)
	flag.Parse()

	if *clientID != "" {
		config.ClientID = *clientID
	}
	if *clientSecret != "" {
		config.ClientSecret = *clientSecret
	}
	if config.ClientID == "" {
		panic("Missing Google Client ID")
	}
	if config.ClientSecret == "" {
		panic("Missing Google Client Secret")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  GOOGLE_CALLBACK_API,
		Endpoint:     googleOAuth2.Endpoint,
		Scopes:       []string{"profile", "email"},
	}
	stateConfig := gologin.DebugOnlyCookieConfig
	router.Handle("/google/login", google.StateHandler(stateConfig, google.LoginHandler(oauth2Config, nil)))
	router.Handle(GOOGLE_CALLBACK_URI, google.StateHandler(stateConfig, google.CallbackHandler(oauth2Config, issueSession(), nil)))

	println("Listen on port " + WEB_PORT)
	http.ListenAndServe(":"+WEB_PORT, router)
}

// index serves the index html file
func index(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user := getUser(session)
	tpl.ExecuteTemplate(w, "index.gohtml", user)
}

// login authenticates the user
func login(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	jsonData := map[string]string{"Email": email, "Password": password}
	jsonValue, _ := json.Marshal(jsonData)
	apiResp, err := http.Post(GetApiUrl("login"), "application/json", bytes.NewBuffer(jsonValue))

	if err == nil {
		data, _ := ioutil.ReadAll(apiResp.Body)
		println(string(data))
		var userResp UserResp
		json.Unmarshal(data, &userResp)

		if userResp.Status >= 400 {
			session.AddFlash("The email or password was incorrect")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}

		userResp.Data.Authenticated = true
		session.Values[SESSION_USER_KEY] = userResp.Data

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

// logout revokes authentication for a user
func logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values[SESSION_USER_KEY] = User{}
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// secret displays the secret message for authorized users
func profile(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := getUser(session)

	if auth := user.Authenticated; !auth {
		session.AddFlash("You don't have access!")
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/forbidden", http.StatusFound)
		return
	}

	if r.Method == "POST" {
		id := user.ID
		address := r.FormValue("address")
		fullname := r.FormValue("fullname")
		telephone := r.FormValue("telephone")
		jsonData := map[string]string{"Fullname": fullname, "Address": address, "Telephone": telephone}
		jsonValue, _ := json.Marshal(jsonData)

		client := &http.Client{}
		body := bytes.NewBuffer(jsonValue)
		req, err := http.NewRequest(http.MethodPut, GetApiUrl("users/"+strconv.Itoa(id)), body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		apiResp, err := client.Do(req)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer apiResp.Body.Close()

		data, _ := ioutil.ReadAll(apiResp.Body)
		println(string(data))

		user.Address = address
		user.Fullname = fullname
		user.Telephone = telephone

		session.Values[SESSION_USER_KEY] = user
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	tpl.ExecuteTemplate(w, "profile.gohtml", user)
}

func register(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")
		address := r.FormValue("address")
		fullname := r.FormValue("fullname")
		telephone := r.FormValue("telephone")
		jsonData := map[string]string{"Email": email, "Password": password, "Fullname": fullname,
			"Address": address, "Telephone": telephone}
		jsonValue, _ := json.Marshal(jsonData)
		response, err := http.Post(GetApiUrl("users"), "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			println("The HTTP request failed with error %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			println(string(data))

			var userResp UserResp
			json.Unmarshal(data, &userResp)

			if userResp.Status >= 400 {
				session.AddFlash("The email is existed.")
				err = session.Save(r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, "/forbidden", http.StatusFound)
				return
			}

			tpl.ExecuteTemplate(w, "register_success.gohtml", nil)
			return
		}
	}
	tpl.ExecuteTemplate(w, "register.gohtml", nil)
	return
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashMessages := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "forbidden.gohtml", flashMessages)
}

func forgot(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		email := r.FormValue("email")

		jsonData := map[string]string{"email": email}
		jsonValue, _ := json.Marshal(jsonData)
		response, err := http.Post(GetApiUrl("users/verify_email"), "application/json", bytes.NewBuffer(jsonValue))

		if err != nil {
			println("The HTTP request failed with error %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			println(string(data))

			var userResp UserResp
			json.Unmarshal(data, &userResp)
			if userResp.Status >= 400 {
				session.AddFlash("The email is not existed.")
				err = session.Save(r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/forbidden", http.StatusFound)
				return
			}

			u := uuid.NewV4()
			jsonData := map[string]interface{}{"ResetKey": u.String()}
			jsonValue, _ = json.Marshal(jsonData)

			client := &http.Client{}
			body := bytes.NewBuffer(jsonValue)
			req, err := http.NewRequest(http.MethodPut, GetApiUrl("users/"+strconv.Itoa(userResp.Data.ID)), body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			apiResp, err := client.Do(req)
			if err != nil {
				println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			data, _ = ioutil.ReadAll(apiResp.Body)
			println(string(data))

			json.Unmarshal(data, &userResp)
			defer apiResp.Body.Close()

			if userResp.Status >= 400 {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sendResetRequest(email, u.String())
			session.AddFlash("You've been sent a reset password link. You must check your email.")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/notice", http.StatusFound)
			return
		}
	}

	tpl.ExecuteTemplate(w, "forgot.gohtml", nil)
}

func notice(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashMessages := session.Flashes()
	err = session.Save(r, w)
	tpl.ExecuteTemplate(w, "notice.gohtml", flashMessages)
	return
}

func verify(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, COOKIE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	jsonData := map[string]string{"ResetKey": uuid}
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(GetApiUrl("users/verify_reset_key"), "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := ioutil.ReadAll(response.Body)
	println(string(data))

	var userResp UserResp
	json.Unmarshal(data, &userResp)
	if userResp.Status >= 400 {
		session.AddFlash("The reset key is not correct.")
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/forbidden", http.StatusFound)
		return
	}

	if r.Method == "POST" {
		password := r.FormValue("password")
		password2 := r.FormValue("password2")
		if password != password2 {
			session.AddFlash("The password does not match with the confirmed password.")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}
		userResp.Data.Password = password
		userResp.Data.ResetKey = ""

		jsonValue, _ = json.Marshal(userResp.Data)

		client := &http.Client{}
		body := bytes.NewBuffer(jsonValue)
		req, err := http.NewRequest(http.MethodPut, GetApiUrl("users/"+strconv.Itoa(userResp.Data.ID)), body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		apiResp, err := client.Do(req)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, _ = ioutil.ReadAll(apiResp.Body)
		println(string(data))

		json.Unmarshal(data, &userResp)
		defer apiResp.Body.Close()

		if userResp.Status >= 400 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.AddFlash("The password changed successful.")
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/notice", http.StatusFound)
		return
	}
	tpl.ExecuteTemplate(w, "reset.gohtml", nil)
	return
}

// getUser returns a user from session s
// on error returns an empty user
func getUser(s *sessions.Session) User {
	val := s.Values[SESSION_USER_KEY]
	var user = User{}
	user, ok := val.(User)
	if !ok {
		return User{Authenticated: false}
	}
	return user
}

func sendResetRequest(email, u string) bool {
	println("Beginning send mail")
	link := "http://localhost:" + WEB_PORT + "/verify/" + u
	host := "smtp.gmail.com"
	port := 587
	msg := gomail.NewMessage()
	msg.SetHeader("From", "huyhuynh.test@gmail.com")
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Password Reset")
	msg.SetBody("text/html", "To reset your password, please click on the link: <a href=\""+link+
		"\">"+link+"</a><br><br>Best Regards,<br>Huy Huynh")
	d := gomail.NewDialer(host, port, "huyhuynh.test@gmail.com", "Just4test")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	s, err := d.Dial()
	if err != nil {
		panic(err)
	}

	if err := gomail.Send(s, msg); err != nil {
		panic(err)
	}
	println("Done send mail")
	return true
}

// issueSession issues a cookie session after successful Google login
func issueSession() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		googleUser, err := google.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Implement a success handler to issue some form of session
		session, err := store.Get(r, COOKIE_NAME)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var user User
		user = User{
			Authenticated: true,
			Fullname:      googleUser.Name,
			Email:         googleUser.Email,
		}
		// user.Authenticated = true
		// user.ID = googleUser.Id
		session.Values[SESSION_USER_KEY] = user
		session.Save(r, w)
		http.Redirect(w, r, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}
