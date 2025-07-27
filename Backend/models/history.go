package models

import (
	"time"

	"gorm.io/gorm"
)

type LoanHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	ExemplarID   uint      `json:"exemplar_id" gorm:"not null"`
	BookID       uint      `json:"book_id" gorm:"not null"`
	LoanDate     time.Time `json:"loan_date" gorm:"not null"`
	DueDate      time.Time `json:"due_date" gorm:"not null"`
	ReturnedDate time.Time `json:"returned_date" gorm:"not null"`
	WasOverdue   bool      `json:"was_overdue" gorm:"default:false"`
	OverdueDays  int       `json:"overdue_days" gorm:"default:0"`

	// Relaciones
	User     User     `json:"user" gorm:"foreignKey:UserID"`
	Exemplar Exemplar `json:"exemplar" gorm:"foreignKey:ExemplarID"`
	Book     Book     `json:"book" gorm:"foreignKey:BookID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type FineHistory struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	StartDate        time.Time `json:"start_date" gorm:"not null"`
	EndDate          time.Time `json:"end_date" gorm:"not null"`
	AccumulatedDays  int       `json:"accumulated_days" gorm:"not null"`
	TotalPenaltyDays int       `json:"total_penalty_days" gorm:"not null"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
