package models

import "github.com/dgrijalva/jwt-go"

var JwtKey = []byte("todo_secret_key")

type User struct {
	Id       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type Task struct {
	Task        string `json:"task" db:"task"`
	IsCompleted bool   `json:"isCompleted" db:"is_task_completed"`
}

type SelectedUser struct {
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
}

type Claims struct {
	UserId      int    `json:"userId"`
	SessionToke string `json:"sessionToken"`
	jwt.StandardClaims
}

type SetStatus struct {
	TaskId          int  `json:"taskId"`
	IsTaskCompleted bool `json:"isTaskCompleted"`
}

type UserInfo struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
