package entity

import "time"

const (
	TaskStatusInitiate   int = 0
	TaskStatusOnProgress int = 1
	TaskStatusDone       int = 2
)

type Task struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      int        `json:"status"`
	Attachment  *string    `json:"attachment"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type Attachment struct {
	ImageURL string `json:"imageUrl"`
}
