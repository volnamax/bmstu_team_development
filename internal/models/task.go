package models

import "github.com/google/uuid"

type TaskBody struct {
	Title       string
	Description string
}

type TaskShortInfo struct {
	ID     uuid.UUID
	IsDone bool
	Title  string
}

type TaskFullInfo struct {
	ID          uuid.UUID
	Title       string
	Description string
	IsDone      bool
	Categories  []Category
}
