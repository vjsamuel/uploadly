package common

import "time"

type Entity struct {
	UploadTime   time.Time `datastore:"upload_time"`
	LastModified time.Time `datastore:"last_modified"`
	Version      int       `datastore:"version"`
	Size         int64     `datastore:"size"`
	Type         string    `datastore:type`
	Description string     `datastore:description`
}

type Profile struct {
	FirstName string `datastore:"first_name"`
	LastName  string `datastore:"last_name"`
}

type Response struct {
	File         string `json:"file"`
	UploadTime   time.Time `json:"upload_time"`
	LastModified time.Time `json:"last_modified"`
	Version      int `json:"version"`
	Size         int64 `json:"size"`
	Type         string `json:"type"`
	Description         string `json:"description"`
}
