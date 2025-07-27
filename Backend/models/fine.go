package models

import (
	"time"

	"gorm.io/gorm"
)

type Fine struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	UserID          uint       `json:"user_id" gorm:"not null"`
	StartDate       time.Time  `json:"start_date" gorm:"not null"`
	AccumulatedDays int        `json:"accumulated_days" gorm:"default:0"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	IsActive        bool       `json:"is_active" gorm:"default:true"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Métodos de negocio
func (f *Fine) CalculateEndDate() time.Time {
	// La multa dura el doble de días de retraso
	penaltyDays := f.AccumulatedDays * 2
	return f.StartDate.AddDate(0, 0, penaltyDays)
}

func (f *Fine) IsExpired() bool {
	if f.EndDate == nil {
		return false
	}
	return time.Now().After(*f.EndDate)
}

func (f *Fine) GetRemainingDays() int {
	if f.EndDate == nil {
		endDate := f.CalculateEndDate()
		f.EndDate = &endDate
	}

	if f.IsExpired() {
		return 0
	}

	return int(f.EndDate.Sub(time.Now()).Hours() / 24)
}
