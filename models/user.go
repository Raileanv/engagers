package models

type User struct {
	Model
	GithubID       int `json:"github_id"`
	Name           string `json:"name"`
	Login string `json:"login"`
	TemporaryToken string `json:"temporary_token"`
	PublicToken    string `json:"public_token"`
	AccessToken    string `json:"access_token"`
	Email          interface{} `gorm:"type:varchar(100);unique_index" json:"email"`
	AvatarUrl      string `json:"avatar_url"`
}
