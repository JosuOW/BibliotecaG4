package models

import (
	"time"

	"gorm.io/gorm"
)

type Loan struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"not null"`
	ExemplarID uint       `json:"exemplar_id" gorm:"not null"`
	LoanDate   time.Time  `json:"loan_date" gorm:"not null"`
	DueDate    time.Time  `json:"due_date" gorm:"not null"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`

	// Relaciones
	User     User     `json:"user" gorm:"foreignKey:UserID"`
	Exemplar Exemplar `json:"exemplar" gorm:"foreignKey:ExemplarID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Métodos de negocio
func (l *Loan) IsOverdue() bool {
	if l.ReturnedAt != nil {
		return false // Ya fue devuelto
	}
	return time.Now().After(l.DueDate)
}

func (l *Loan) GetOverdueDays() int {
	if !l.IsOverdue() {
		return 0
	}
	return int(time.Since(l.DueDate).Hours() / 24)
}

func (l *Loan) IsActive() bool {
	return l.ReturnedAt == nil
}

// Request/Response DTOs
type LoanRequest struct {
	ExemplarID uint `json:"exemplar_id" validate:"required"`
}

type LoanResponse struct {
	ID           uint       `json:"id"`
	BookTitle    string     `json:"book_title"`
	BookAuthor   string     `json:"book_author"`
	ExemplarCode string     `json:"exemplar_code"`
	LoanDate     time.Time  `json:"loan_date"`
	DueDate      time.Time  `json:"due_date"`
	ReturnedAt   *time.Time `json:"returned_at,omitempty"`
	IsOverdue    bool       `json:"is_overdue"`
	OverdueDays  int        `json:"overdue_days"`
}
