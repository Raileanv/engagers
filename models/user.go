package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	GithubID       int
	Name           string
	TemporaryToken string
	PublicToken    string
	AccessToken    string
	Email          interface{} `gorm:"type:varchar(100);unique_index" json:"email"`
	AvatarUrl      string
}
