package model

import "time"

type Broker struct {
	ID        int64     `json:"id"         gorm:"column:id;primaryKey"`
	Name      string    `json:"name"       gorm:"column:name"`
	Secret    string    `json:"secret"     gorm:"column:secret"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName gorm table name
func (Broker) TableName() string {
	return "broker"
}
