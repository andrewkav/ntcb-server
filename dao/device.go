package dao

import "time"

type Device struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

func (Device) TableName() string {
	return "device"
}
