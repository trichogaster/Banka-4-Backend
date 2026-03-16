package model

import (
	"time"
)

type AccountType string
type AccountKind string
type Subtype string

const (
	AccountTypePersonal AccountType = "Personal"
	AccountTypeBusiness AccountType = "Business"
)

const (
	AccountKindCurrent AccountKind = "Current"
	AccountKindForeign AccountKind = "Foreign"
)

const (
	SubtypeStandard   Subtype = "Standard"
	SubtypeSavings    Subtype = "Savings"
	SubtypePension    Subtype = "Pension"
	SubtypeYouth      Subtype = "Youth"
	SubtypeStudent    Subtype = "Student"
	SubtypeUnemployed Subtype = "Unemployed"
	SubtypeLLC        Subtype = "LLC"
	SubtypeJointStock Subtype = "JointStock"
	SubtypeFoundation Subtype = "Foundation"
)

var AllowedForeignCurrencies = map[string]bool{
	"EUR": true, "CHF": true, "USD": true,
	"GBP": true, "JPY": true, "CAD": true, "AUD": true,
}

type Account struct {
	AccountNumber    string      `gorm:"primaryKey;size:18"`
	Name             string
	ClientID         uint        `gorm:"not null;index"`

	CompanyID        *uint       `gorm:"index"`
	Company          *Company

	EmployeeID       uint        `gorm:"not null"`

	CurrencyID       uint        `gorm:"index;not null"`
	Currency         Currency

	Balance          float64     `gorm:"not null;default:0"`
	AvailableBalance float64     `gorm:"not null;default:0"`

	CreatedAt        time.Time   `gorm:"autoCreateTime"`
	ExpiresAt        time.Time   `gorm:"not null"`

	Status           string      `gorm:"not null;default:Active"`

	AccountType      AccountType `gorm:"not null;size:20"`
	AccountKind      AccountKind `gorm:"not null;size:20"`
	Subtype          Subtype     `gorm:"size:20"`

	MaintenanceFee   float64     `gorm:"not null;default:0"`
	DailyLimit       float64     `gorm:"not null;default:0"`
	MonthlyLimit     float64     `gorm:"not null;default:0"`
	DailySpending    float64     `gorm:"not null;default:0"`
	MonthlySpending  float64     `gorm:"not null;default:0"`
}
