package model

import "time"

// Status zahteva za kredit (enum)
type LoanRequestStatus string

const (
	LoanRequestPending  LoanRequestStatus = "PENDING"
	LoanRequestApproved LoanRequestStatus = "APPROVED"
	LoanRequestRejected LoanRequestStatus = "REJECTED"
)

// Tip kredita koji banka nudi (npr. Kes kredit, Stambeni kredit)
type LoanType struct {
	LoanTypeID         uint   `gorm:"primaryKey"`
	Name               string `gorm:"size:100;not null"`
	Description        string
	BankMargin         float64 // Marza banke u procentima (npr. 2.5%)
	BaseInterestRate   float64 // Bazna kamatna stopa (npr. 3.0%)
	MinRepaymentPeriod int     // Minimalni broj meseci
	MaxRepaymentPeriod int     // Maksimalni broj meseci
}

// Zahtev klijenta za kredit
type LoanRequest struct {
	ID                 uint              `gorm:"primaryKey"`
	ClientID           uint              `gorm:"not null"`         // Veza ka klijentu (User)
	AccountNumber      string            `gorm:"size:20;not null"` // Racun na koji se prebacuju pare
	LoanTypeID         uint              `gorm:"not null"`         // Veza ka tipu kredita
	Amount             float64           `gorm:"not null"`         // Iznos koji klijent trazi
	RepaymentPeriod    int               `gorm:"not null"`         // Na koliko meseci (n)
	CalculatedRate     float64           // Kamata (Marza + Osnovna kamata)
	MonthlyInstallment float64           // Izracunata mesecna rata (A)
	Status             LoanRequestStatus `gorm:"size:20;default:'PENDING'"`
	CreatedAt          time.Time

	LoanType LoanType
}

// Status aktivnog kredita
type LoanStatus string

const (
	LoanStatusActive    LoanStatus = "ACTIVE"
	LoanStatusCompleted LoanStatus = "COMPLETED"
	LoanStatusDefault   LoanStatus = "DEFAULT"
)

// Status pojedinacne rate kredita
type InstallmentStatus string

const (
	InstallmentStatusPending  InstallmentStatus = "PENDING"
	InstallmentStatusPaid     InstallmentStatus = "PAID"
	InstallmentStatusUnpaid   InstallmentStatus = "UNPAID"
	InstallmentStatusRetrying InstallmentStatus = "RETRYING"
)

// Loan je aktivni kredit — kreira se kada se LoanRequest odobri
type Loan struct {
	ID            uint `gorm:"primaryKey"`
	LoanRequestID uint `gorm:"not null;uniqueIndex"` // 1:1 veza ka zahtevu
	LoanRequest   LoanRequest
	AccountNumber string `gorm:"size:20;not null"` // racun klijenta na koji je isplacen kredit
	ClientID      uint   `gorm:"not null;index"`

	TotalAmount        float64 `gorm:"not null"` // originalni iznos kredita
	RemainingDebt      float64 `gorm:"not null"` // preostali dug
	MonthlyInstallment float64 `gorm:"not null"` // trenutna mesecna rata
	InterestRate       float64 `gorm:"not null"` // trenutna kamatna stopa
	IsVariableRate     bool    `gorm:"not null;default:false"`

	RepaymentPeriod  int `gorm:"not null"` // ukupan broj meseci
	PaidInstallments int `gorm:"not null;default:0"`

	StartDate           time.Time `gorm:"not null"`
	NextInstallmentDate time.Time `gorm:"not null;index"` // index za brzo pronalazenje dospelih rata

	Status LoanStatus `gorm:"size:20;not null;default:'ACTIVE'"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Installments []LoanInstallment
}

// LoanInstallment predstavlja jednu mesecnu ratu kredita
type LoanInstallment struct {
	ID     uint `gorm:"primaryKey"`
	LoanID uint `gorm:"not null;index"`
	Loan   Loan

	InstallmentNumber int     `gorm:"not null"` // redni broj rate (1, 2, 3...)
	Amount            float64 `gorm:"not null"`
	InterestRate      float64 `gorm:"not null"` // kamata koja je vazila u trenutku kreiranja rate

	DueDate time.Time  `gorm:"not null;index"`
	PaidAt  *time.Time // null ako rata nije placena
	RetryAt *time.Time // vreme sledeceg pokusaja naplate, null ako nije u retry statusu

	Status InstallmentStatus `gorm:"size:20;not null;default:'PENDING'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
