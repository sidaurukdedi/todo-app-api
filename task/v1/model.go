package task

import (
	"io"
	"time"
)

type TaskResponse struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Status      *int       `json:"status"`
	Attachment  *string    `json:"attachment"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

type TaskRequest struct {
	Name        string     `json:"name" validate:"required"`
	Description *string    `json:"description" validate:"-"`
	Status      *int       `json:"status" validate:"-"`
	Attachment  *string    `json:"attachment" validate:"-"`
	CreatedAt   time.Time  `json:"createdAt" validate:"-"`
	UpdatedAt   *time.Time `json:"updatedAt" validate:"-"`
}

type UploadAttachmentRequest struct {
	Attachment struct {
		File          io.Reader `validate:"-"`
		Size          int64     `validate:"-"`
		FileName      string    `validate:"-"`
		FileExtension string    `validate:"-"`
		FileNameParam string    `validate:"-"`
	} `validate:"-"`
}
