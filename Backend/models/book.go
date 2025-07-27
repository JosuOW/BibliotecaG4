package models

import "time"

type Book struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	ISBN               string `json:"isbn" gorm:"unique;not null" validate:"required"`
	Title              string `json:"title" gorm:"not null" validate:"required"`
	Author             string `json:"author" gorm:"not null" validate:"required"`
	Pages              int    `json:"pages" validate:"min=1"`
	TotalExemplars     int    `json:"total_exemplars" gorm:"default:0"`
	AvailableExemplars int    `json:"available_exemplars" gorm:"default:0"`
	CoverURL           string `json:"cover_url"`

	// Relaciones
	Exemplars       []Exemplar       `json:"exemplars,omitempty" gorm:"foreignKey:BookID"`
	Recommendations []Recommendation `json:"recommendations,omitempty" gorm:"foreignKey:OriginBookID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Exemplar struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	BookID          uint      `json:"book_id" gorm:"not null"`
	Code            string    `json:"code" gorm:"unique;not null" validate:"required"`
	AcquisitionDate time.Time `json:"acquisition_date"`
	Observations    string    `json:"observations"`
	IsAvailable     bool      `json:"is_available" gorm:"default:true"`

	// Relaciones
	Book  Book   `json:"book,omitempty" gorm:"foreignKey:BookID"`
	Loans []Loan `json:"loans,omitempty" gorm:"foreignKey:ExemplarID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Recommendation struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	OriginBookID uint   `json:"origin_book_id" gorm:"not null"`
	TargetBookID uint   `json:"target_book_id" gorm:"not null"`
	Comment      string `json:"comment"`

	// Relaciones
	OriginBook Book `json:"origin_book,omitempty" gorm:"foreignKey:OriginBookID"`
	TargetBook Book `json:"target_book,omitempty" gorm:"foreignKey:TargetBookID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
