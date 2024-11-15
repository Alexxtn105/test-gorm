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

// region Структуры для примера "один к одному" и "один ко многим"

type User struct {
	gorm.Model
	ID         uint64      `gorm:"primaryKey"`
	Username   string      `gorm:"size:64"`
	Password   string      `gorm:"size:255"`
	Notes      []Note      // Ассоциация с записями. В запросе используем DB.Preload("Notes") для сокращения вызовов
	CreditCard *CreditCard // Ассоциация с кредитными картами. Используем указатель, чтобы избежать цикличности, так как в структуре User будет ссылка на CreditCard В запросе используем DB.Preload("CreditCard") для сокращения вызовов
}

type Note struct {
	gorm.Model
	ID      uint64 `gorm:"primaryKey"`
	Name    string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	UserID  uint64 `gorm:"index"`
	User    User   // Ассоциация с пользователем. В запросе используем DB.Preload("User") для сокращения вызовов
}

type CreditCard struct {
	gorm.Model
	ID     uint64 `gorm:"primaryKey"`
	Number string `gorm:"size:50"`
	UserID uint64 `gorm:"index"`
}

// endregion

// region Структуры для примера "Многие ко многим"

/* Пример данных
"Iron Man"				Robert Downey Jr.
"Avengers"				Robert Downey Jr., Chris Evans, Scarlett Johansson
"Avengers Infinity War"	Robert Downey Jr., Chris Evans, Scarlett Johansson, Chadwick Boseman
"Sherlock Holmes"		Robert Downey Jr.
"Lost in Translation"	Scarlett Johansson
"Marriage Story"		Scarlett Johansson
*/

type Movie struct {
	gorm.Model
	Name string
}

type Actor struct {
	gorm.Model
	Name string
}

type Filmography struct {
	gorm.Model
	MovieID int
	ActorID int
}

func (Filmography) TableName() string {
	return "filmography"
}

//endregion

//region Подключение к БД и миграция

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
		&Movie{},
		&Actor{},
		&Filmography{},
	)
}

//endregion

func main() {
	connectDatabase()
	dbMigrate()

	//region //Вариант без ассоциаций (отношение "один к одному" и "один ко многим")
	//var note Note
	//DB.First(&note)
	//var user User
	//DB.Where("ID = ?", note.UserID).First(&user)
	//fmt.Printf("User from a note: %s\n", user.Username)
	//fmt.Println("-----------------------------")
	//
	//var notes []Note
	//DB.Where("user_id = ?", user.ID).Find(&notes)
	//
	//fmt.Printf("Notes from a user:\n")
	//for _, element := range notes {
	//	fmt.Printf("%s - %s\n", element.Name, element.Content)
	//}
	//fmt.Println("-----------------------------")
	//
	//var cc CreditCard
	//DB.Where("user_id = ?", user.ID).First(&cc)
	//fmt.Printf("Credit card from a user: %s\n", cc.Number)
	//endregion

	//region Вариант с ассоциациями (отношение "один к одному" и "один ко многим")
	var note Note
	DB.Preload("User").First(&note) // предварительно загружаем талицу пользователей - сможем использовать далее в виде "note.User"
	fmt.Printf("User from a note: %s\n", note.User.Username)
	fmt.Println("-----------------------------")

	var user User
	DB.Preload("Notes").Preload("CreditCard").Where("username = ?", "Alex").First(&user)
	fmt.Printf("Notes from a user:\n")
	for _, element := range user.Notes {
		fmt.Printf("%s - %s\n", element.Name, element.Content)
	}
	fmt.Println("-----------------------------")

	var cc CreditCard
	DB.Where("user_id = ?", user.ID).First(&cc)
	fmt.Printf("Credit card from a user: %s\n", user.CreditCard.Number)
	//endregion

	//region Вариант для отношения "Многие ко многим"
	var movie Movie
	DB.Where("name = ?", "Avengers Infinity War").First(&movie)
	fmt.Printf("Movie: %s", movie.Name)

	var filmography []Filmography
	DB.Where("movie_id = ?", movie.ID).Find(&filmography)
	fmt.Printf("Filmography count: %v\n\n", len(filmography))

	var actor_ids []int
	for _, element := range filmography {
		actor_ids = append(actor_ids, element.ActorID)
	}
	fmt.Printf("Actor IDs: %v\n\n", actor_ids)

	var actors []Actor
	DB.Where("id IN ?", actor_ids).Find(&actors)
	fmt.Println("Actors:")
	for _, actor := range actors {
		fmt.Printf("%s\n", actor.Name)
	}
	//endregion
}
