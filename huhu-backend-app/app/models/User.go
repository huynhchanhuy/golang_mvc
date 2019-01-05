package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type User struct {
	gorm.Model
	Email     string
	Password  string
	Fullname  string
	Address   string
	Telephone string
	ResetKey  string
	// CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	// UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	// DeletedAt time.Time `gorm:"column:deleted_at" json:"deleted_at"`
	// Category string `gorm:"column:category" json:"category"`
}

func (u *User) TableName() string {
	return "users"
}

func InsertUser(db *gorm.DB, u *User) (err error) {
	if err = db.Save(u).Error; err != nil {
		return err
	}
	return nil
}

func GetAllUser(db *gorm.DB, u *[]User) (err error) {
	if err = db.Order("id desc").Find(u).Error; err != nil {
		return err
	}
	return nil
}

func OneUserGetting(db *gorm.DB, ids int, u *User) (err error) {
	if err := db.Where("id = ?", ids).First(&u).Error; err != nil {
		return err
	}
	return nil
}

func UpdateUser(db *gorm.DB, u *User) (err error) {
	if err = db.Save(u).Error; err != nil {
		return err
	}
	return nil
}

func DeletedUser(db *gorm.DB, u *User) (err error) {
	if err = db.Delete(u).Error; err != nil {
		return err
	}
	return nil
}

func OneUserLogin(db *gorm.DB, email string, u *User) (err error) {
	if err := db.Where("email = ?", email).First(&u).Error; err != nil {
		return err
	}
	return nil
}

func OneUserResetKey(db *gorm.DB, resetKey string, u *User) (err error) {
	if err := db.Where("reset_key = ?", resetKey).First(&u).Error; err != nil {
		return err
	}
	return nil
}
