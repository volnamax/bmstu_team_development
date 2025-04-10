package models

import "github.com/google/uuid"

type TaskBody struct {
	Title       string
	Description string
}

type TaskMeta struct {
	Id     uuid.UUID
	IsDone bool
}

type Task struct {
	TaskMeta
	TaskBody
}
