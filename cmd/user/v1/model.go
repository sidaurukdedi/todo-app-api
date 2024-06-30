package user

import "time"

type UserResponse struct {
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
