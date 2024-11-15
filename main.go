package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func main() {

	connectDatabase()
	dbMigrate()

	var note Note
	DB.First(&note)

	var user User
	DB.Where("ID = ?", note.UserID).First(&user)
	fmt.Printf("User from a note: %s\n", user.Username)
	fmt.Println("-----------------------------")

	var notes []Note
	DB.Where("user_id = ?", user.ID).Find(&notes)

	fmt.Printf("Notes from a user:\n")
	for _, element := range notes {
		fmt.Printf("%s - %s\n", element.Name, element.Content)
	}
	fmt.Println("-----------------------------")

	var cc CreditCard
	DB.Where("user_id = ?", user.ID).First(&cc)
	fmt.Printf("Credit card from a user: %s\n", cc.Number)

}

type User struct {
	gorm.Model
	ID       uint64 `gorm:"primaryKey"`
	Username string `gorm:"size:64"`
	Password string `gorm:"size:255"`
}

type Note struct {
	gorm.Model
	ID      uint64 `gorm:"primaryKey"`
	Name    string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	UserID  uint64 `gorm:"index"`
}

type CreditCard struct {
	gorm.Model
	ID     uint64 `gorm:"primaryKey"`
	Number string `gorm:"size:50"`
	UserID uint64 `gorm:"index"`
}

var DB *gorm.DB

func connectDatabase() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, //
			LogLevel:                  logger.Info, // уровень логирования
			IgnoreRecordNotFoundError: true,        // игнорировать ErrRecordNotFound для логгера
			Colorful:                  true,        // расцветка
		},
	)

	database, err := gorm.Open(sqlite.Open("storage/storage.db"), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic("Failed to connect DB!")
	}

	DB = database
}

func dbMigrate() {
	DB.AutoMigrate(
		&Note{},
		&User{},
		&CreditCard{},
	)
}
