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
	//Добавим поле для ассоциаций gorm
	Actors []Actor `gorm:"many2many:filmography;"` // Задаем отношение "многие ко многим" через таблицу filmography. В этом случае создания структуры Filmography не требуется
}

type Actor struct {
	gorm.Model
	Name   string
	Movies []Movie `gorm:"many2many:filmography;"` // Задаем отношение "многие ко многим" через таблицу filmography. В этом случае создания структуры Filmography не требуется
}

// Эта структура не нужна, поскольку в структуре Movie задана ассоциация Actors вида: Actors []Actor `gorm:"many2many:filmography;"`
//type Filmography struct {
//	gorm.Model
//	MovieID int
//	ActorID int
//}
//
//func (Filmography) TableName() string {
//	return "filmography"
//}

//endregion

//region Структуры для Gorm Scopes

type Consumer struct {
	gorm.Model
	ID     int64   `gorm:"primaryKey"`
	Name   string  `gorm:"size:255"`
	Email  string  `gorm:"size:255"`
	Orders []Order // У одного пользователя может быть несколько заказов
}

type Order struct {
	gorm.Model
	ID          int64 `gorm:"primaryKey"`
	ConsumerID  int64
	OrderTime   time.Time // Время заказа
	PaymentMode string    `gorm:"size:255"` // Card or Cache
	Price       int       // Стоимость
	Consumer    Consumer  // Пользователь, осуществивший заказ
}

//endregion

// region Gorm Scopes

// CardOrders функция, использующая gorm scope и возращающая все заказы, оплаченные картой
func CardOrders(db *gorm.DB) *gorm.DB {
	// РЕГИСТР В УСЛОВИИ ПОИСКА ИМЕЕТ ЗНАЧЕНИЕ!!!
	return db.Where("payment_mode = ?", "Card")
}

// PriceGreaterThan30 получить все заказы стоимостью более 30
func PriceGreaterThan30(db *gorm.DB) *gorm.DB {
	return db.Where("price > ?", 30)
}

// UsersFromDomain - возвращает всех пользователей.
// В качестве возвращаемого значения используется функция-замыкание с сигнатурой gorm Scope: func(db *gorm.DB) *gorm.DB
func UsersFromDomain(domain string) func(db *gorm.DB) *gorm.DB {
	// возвращаем функцию-замыкание
	return func(db *gorm.DB) *gorm.DB {
		// возвращаем всех пользователей, чья почта заканчивается на указанный в параметре domain (например ".com")
		return db.Where("email like ?", "%"+domain)
	}
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
		&Consumer{},
		&Order{},
		//&Filmography{}, // не нужно, gorm создаст из ассоциаций
	)
}

//endregion

//region main

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
	// необходимо сделать Preload таблицы Actors (для работы ассоциаций gorm)
	DB.Where("name = ?", "Avengers Infinity War").Preload("Actors").First(&movie)

	// Код ниже не нужен, поскольку мы используем ассоциации gorm и Preload
	//fmt.Printf("Movie: %s", movie.Name)
	//
	//var filmography []Filmography
	//DB.Where("movie_id = ?", movie.ID).Find(&filmography)
	//fmt.Printf("Filmography count: %v\n\n", len(filmography))
	//
	//var actor_ids []int
	//for _, element := range filmography {
	//	actor_ids = append(actor_ids, element.ActorID)
	//}
	//fmt.Printf("Actor IDs: %v\n\n", actor_ids)
	//
	//var actors []Actor
	//DB.Where("id IN ?", actor_ids).Find(&actors)
	fmt.Println("Actors:")
	for _, actor := range movie.Actors {
		fmt.Printf("%s\n", actor.Name)
	}

	var actor Actor
	DB.Where("name = ?", "Robert Downey Jr.").Preload("Movies").First(&actor)
	fmt.Println("Actor: " + actor.Name)

	fmt.Println("Movies:")
	for _, mov := range actor.Movies {
		fmt.Printf("%s\n", mov.Name)
	}
	//endregion

	//region Вариант Gorm Scopes

	// Берем все заказы, оплаченные картой с использованием Scopes.
	// Метод Scopes принимает в качестве аргументов функции CardOrders и PriceGreaterThan30 - они объединяются по условию "AND"

	var orders []Order
	DB.Scopes(CardOrders, PriceGreaterThan30).Find(&orders)
	fmt.Println("orders:")
	for _, order := range orders {
		fmt.Printf("%s: by %s,  %d р.\n", order.OrderTime, order.PaymentMode, order.Price)
	}

	var consumers []Consumer
	domain := ".com"
	// ВАРИАНТ 1 - Обычный вызов
	//DB.Scopes(UsersFromDomain(domain)).Find(&consumers)

	//ВАРИАНТ 2 - С доп. условиями - Используем Preload таблицы Orders, там же фильтруем с использованием scope.
	// Например, выведем всех пользователей из домена .com (UsersFromDomain), оплативших заказы картой (CardOrders)
	DB.Scopes(UsersFromDomain(domain)).Preload("Orders", CardOrders).Find(&consumers)
	fmt.Printf("Consumers with domain %s:\n", domain)
	for _, consumer := range consumers {
		fmt.Printf("%s: %s\n", consumer.Name, consumer.Email)
	}

	// выводим заказ только ПЕРВОГО пользователя
	if len(consumers) > 0 {

		fmt.Printf("Orders of user %s:\n", consumers[0].Name)
		for _, order := range consumers[0].Orders {
			fmt.Printf("%s: by %s,  %d р.\n", order.OrderTime, order.PaymentMode, order.Price)
		}
	}

	//endregion
}

//endregion
