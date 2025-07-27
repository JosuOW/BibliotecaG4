package models

import (
	"time"

	"gorm.io/gorm"
)

type UserType string

const (
	STUDENT UserType = "STUDENT"
	TEACHER UserType = "TEACHER"
	ADMIN   UserType = "ADMIN"
)

type UserStatus string

const (
	ACTIVE UserStatus = "ACTIVE"
	DEBTOR UserStatus = "DEBTOR"
	FINED  UserStatus = "FINED"
)

type User struct {
	ID       uint       `json:"id" gorm:"primaryKey"`
	Login    string     `json:"login" gorm:"unique;not null" validate:"required"`
	Name     string     `json:"name" gorm:"not null" validate:"required"`
	LastName string     `json:"last_name" gorm:"not null" validate:"required"`
	Email    string     `json:"email" gorm:"unique;not null" validate:"required,email"`
	UserType UserType   `json:"user_type" gorm:"not null" validate:"required"`
	Status   UserStatus `json:"status" gorm:"default:ACTIVE"`

	// Dirección
	Street     string `json:"street"`
	Number     string `json:"number"`
	Floor      string `json:"floor"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`

	// Campos específicos por tipo
	ParentsPhone   *string `json:"parents_phone,omitempty"`   // Solo estudiantes
	DepartmentName *string `json:"department_name,omitempty"` // Solo profesores

	// Relaciones
	Loans       []Loan        `json:"loans,omitempty" gorm:"foreignKey:UserID"`
	Fine        *Fine         `json:"fine,omitempty" gorm:"foreignKey:UserID"`
	LoanHistory []LoanHistory `json:"loan_history,omitempty" gorm:"foreignKey:UserID"`
	FineHistory []FineHistory `json:"fine_history,omitempty" gorm:"foreignKey:UserID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Métodos de negocio
func (u *User) GetMaxLoans() int {
	switch u.UserType {
	case STUDENT:
		return 5
	case TEACHER:
		return 8
	default:
		return 0
	}
}

func (u *User) GetLoanDays() int {
	switch u.UserType {
	case STUDENT:
		return 7
	case TEACHER:
		return 30
	default:
		return 0
	}
}

func (u *User) CanBorrow() bool {
	return u.Status == ACTIVE
}

func (u *User) GetCurrentLoansCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&Loan{}).Where("user_id = ? AND returned_at IS NULL", u.ID).Count(&count)
	return count
}
