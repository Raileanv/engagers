package main

import (
	"context"
	"encoding/json"
	"engagers/models"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type OToken struct {
	*oauth2.Token
	TemporaryToken string
	PublicToken    string
}

type GithubUserInfo struct {
	Login             string      `json:"login"`
	ID                int         `json:"id"`
	NodeID            string      `json:"node_id"`
	AvatarURL         string      `json:"avatar_url"`
	GravatarID        string      `json:"gravatar_id"`
	URL               string      `json:"url"`
	HTMLURL           string      `json:"html_url"`
	FollowersURL      string      `json:"followers_url"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	OrganizationsURL  string      `json:"organizations_url"`
	ReposURL          string      `json:"repos_url"`
	EventsURL         string      `json:"events_url"`
	ReceivedEventsURL string      `json:"received_events_url"`
	Type              string      `json:"type"`
	SiteAdmin         bool        `json:"site_admin"`
	Name              string      `json:"name"`
	Company           string      `json:"company"`
	Blog              string      `json:"blog"`
	Location          string      `json:"location"`
	Email             interface{} `json:"email"`
	Hireable          bool        `json:"hireable"`
	Bio               string      `json:"bio"`
	PublicRepos       int         `json:"public_repos"`
	PublicGists       int         `json:"public_gists"`
	Followers         int         `json:"followers"`
	Following         int         `json:"following"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

func AuthWithTempTokenHandler(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()

	temporaryToken := keys.Get("temporary_token")

	models.FindUserByTempToken(temporaryToken)
	if models.IsCurrentUserPresent() {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}
	var oAuthToken oauth2.Token
	token := OToken{Token: &oAuthToken, PublicToken: fmt.Sprint(uuid.New()), TemporaryToken: ""}

	UpdateUserToken(&models.CurrentUser, &token)

	fmt.Fprintf(w, models.CurrentUser.PublicToken)
}

func newGithubConfig() oauth2.Config {

	redirectUrl := fmt.Sprintf("%v%v", os.Getenv("BASE_URL"), "users/auth/github/callback")
	return oauth2.Config{
		//ClientID:     "d535e0f5cad826235ff6",
		//ClientSecret: "98ff1eeae3fa9dc64fcb31ecbc302ec3fe0ff0d5",
		//RedirectURL: "http://localhost:3000/users/auth/github/callback",
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  redirectUrl,
		Endpoint: github.Endpoint,
		Scopes:   []string{"user"},
	}
}

var (
	githubConfig = newGithubConfig()
)

func AuthWithGithubHandler(w http.ResponseWriter, r *http.Request) {
	state := RandToken(48)
	authorizationURL := githubConfig.AuthCodeURL(state)

	http.Redirect(w, r, authorizationURL, 301)
}

func AuthGithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	authorizationCode := r.URL.Query().Get("code")

	ck, err := r.Cookie("state")
	if err == nil && (r.URL.Query().Get("state") != ck.Value) {
		fmt.Fprintf(w, "Error: State is not the same")
	}
	oAuthToken, err := githubConfig.Exchange(context.Background(), authorizationCode)
	if err != nil {
		panic(err)
	}

	meRequest, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}
	tokenStr := fmt.Sprint("Token ", oAuthToken.AccessToken)
	meRequest.Header.Set("Authorization", tokenStr)

	meResonse, err := client.Do(meRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	defer meResonse.Body.Close()

	var githubUserInfo GithubUserInfo
	bodyByte, _ := ioutil.ReadAll(meResonse.Body)
	err = json.Unmarshal(bodyByte, &githubUserInfo)
	if err != nil {
		fmt.Println("Unable to unmarshal user info")
	}

	fmt.Println(string(bodyByte))

	token := OToken{Token: oAuthToken, PublicToken: "", TemporaryToken: fmt.Sprint(uuid.New())}
	FindOrCreateUser(&token, &githubUserInfo)

	if models.IsCurrentUserPresent() {
		url := fmt.Sprintf("%v%v", os.Getenv("BASE_URL"), "auth")
		http.Redirect(w, r, url, http.StatusFound)
	}
	tempUrl := fmt.Sprintf("%v%v", os.Getenv("BASE_URL"), "temp_url_handler")
	tempTokenURL := models.GenerateTempTokenUrl(models.CurrentUser.TemporaryToken, tempUrl)
	http.Redirect(w, r, tempTokenURL, 301)
}

func FindOrCreateUser(token *OToken, userInfo *GithubUserInfo) models.User {
	user := models.User{}

	DB.Find(&user, "github_id = ?", userInfo.ID)
	if (models.User{} != user) {
		UpdateUserToken(&user, token)
		models.SetCurrentUser(&user)
		return user
	}

	CreateUser(&user, token, userInfo)
	models.SetCurrentUser(&user)

	return user
}

func UpdateUserToken(user *models.User, token *OToken) {
	if token.AccessToken != "" {
		user.AccessToken = token.AccessToken
	}
	user.TemporaryToken = token.TemporaryToken
	user.PublicToken = token.PublicToken

	DB.Save(&user)
}

func CreateUser(user *models.User, t *OToken, ui *GithubUserInfo) {
	user.AccessToken = t.AccessToken
	user.TemporaryToken = t.TemporaryToken
	user.PublicToken = t.PublicToken
	user.Email = ui.Email
	user.Name = ui.Name
	user.GithubID = ui.ID
	user.AvatarUrl = ui.AvatarURL

	DB.Create(&user)
}
