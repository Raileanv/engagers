package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"os"

	"engagers/models"
)

var (
	timeout = time.Duration(5 * time.Second)
	client  = http.Client{
		Timeout: timeout,
	}

	upgrader = websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 5 * time.Second,
	}
	DB = models.InitDB()
)

func getMeHandler(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	user := models.FindUserByPubToken(reqToken)

	if reqToken == "" || (user == models.User{}) || (user.AccessToken == "") {
		url := fmt.Sprintf("%v%v", os.Getenv("BASE_URL"), "auth_with_github")
		http.Redirect(w, r, url, http.StatusFound)
	}

	getMeUrl := models.GenerateGetMeUrl(user.AccessToken)

	request, _ := http.NewRequest("GET", getMeUrl, nil)
	response, err := client.Do(request)
	if response.StatusCode != 200 {
		fmt.Errorf("ERROR: %s", "Dude try to auth again")
		return
	}
	defer response.Body.Close()

	meBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Errorf("ERROR: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(meBytes)

}

func webSocketsHandler(hub *Hub, w http.ResponseWriter, r *http.Request, params martini.Params) {
	id, _ := strconv.ParseInt(params["session_id"], 10, 32)

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: hub, conn: conn, sessionId: int(id), send: make(chan []byte, 256)}
	client.hub.register <- client

	client.send <- []byte("ХУЯК!!! ТЫ ПОДКЛЮЧИЛСЯ!!")

	go client.readPump()
	go client.writePump()
}

func authChecker(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	models.FindUserByPubToken(reqToken)

	if reqToken == "" || (models.CurrentUser == models.User{}) || (models.CurrentUser.AccessToken == "") {
		url := fmt.Sprintf("%v%v", os.Getenv("BASE_URL"), "auth_with_github")
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func main() {
	defer DB.Close()
	hub := newHub()
	go hub.run()

	m := martini.Classic()

	m.Group("/api/v1", func(r martini.Router) {
		r.Get("/get_me", getMeHandler)

		r.Group("/presentations", func(rr martini.Router) {
			rr.Post("/", models.CreatePresentationHandler)
			rr.Post("/:presentation_id/session", func(w http.ResponseWriter, r *http.Request, params martini.Params) {
				models.PostAddSessionToPresentation(w, r, params)
			})
			rr.Post("/:presentation_id/quiz", func(w http.ResponseWriter, r *http.Request, params martini.Params) {
				models.PostAddQuizToPresentation(w, r, params)
			})
			rr.Get("/:presentation_id", func(w http.ResponseWriter, r *http.Request, params martini.Params) {
				models.GetPresentationHandler(w, r, params)
			})
			rr.Get("/:presentation_id/sessions", func(w http.ResponseWriter, r *http.Request, params martini.Params) {
				models.GetPresentationSessionsHandler(w, r, params)
			})
			rr.Get("/", models.GetPresentationsHandler)
		})

		r.Group("/conference", func(rr martini.Router) {
			rr.Post("/", models.CreateConferenceHandler)
			rr.Get("/", models.GetConferencesHandler)
			rr.Get("/:conference_id", func(w http.ResponseWriter, r *http.Request, params martini.Params) {
				models.GetConferenceHandler(w, r, params)
			})
		})

	}, authChecker)

	m.Get("/ws/:session_id", func(w http.ResponseWriter, r *http.Request, p martini.Params) {
		webSocketsHandler(hub, w, r, p)
	}, authChecker)

	m.Get("/users/auth/github/callback", AuthGithubCallbackHandler)
	m.Post("/auth_with_temporary_token", AuthWithTempTokenHandler)
	m.Get("/auth_with_github", AuthWithGithubHandler)

	m.Run()
}
