package migrate

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type User struct {
	gorm.Model
	// Id        int    `gorm:"column:id"`
	Email     string `gorm:"column:email" json:"email"`
	Password  string `gorm:"column:password" json:"password"`
	Fullname  string `gorm:"column:fullname" json:"fullname"`
	Address   string `gorm:"column:address" json:"address"`
	Telephone string `gorm:"column:telephone" json:"telephone"`
	ResetKey  string `gorm:"column:reset_key" json:"reset_key"`
	// Category string `gorm:"column:category" json:"category"`
}

func DBMigrate(db *gorm.DB) *gorm.DB {
	db.AutoMigrate(&User{})
	hasUser := db.HasTable(&User{})
	fmt.Println("Table user is", hasUser)
	if !hasUser {
		db.CreateTable(&User{})
	}

	return db
}
