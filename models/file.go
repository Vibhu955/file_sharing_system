package models

import "time"

type File struct {
	ID        int       `json:"id"`
	UserEmail string    `json:"user_email"`
	FileName  string    `json:"file_name"`
	Size      int64     `json:"size"`
	URL       string    `json:"url"`
	UploadAt  time.Time `json:"upload_at"`
}
