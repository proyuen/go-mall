package model

type User struct {
	Base
	Username     string `gorm:"uniqueIndex;not null;type:varchar(50)" json:"username"`
	PasswordHash string `gorm:"not null;type:varchar(255)" json:"-"`
	Email        string `gorm:"uniqueIndex;not null;type:varchar(100)" json:"email"`
	Role         string `gorm:"default:'user';type:varchar(20)" json:"role"`
}
