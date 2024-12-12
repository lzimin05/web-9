package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "lenya"
	password = "111222qqq"
	dbname   = "sandbox"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

type CountRequest struct {
	Msg string `json:"msg" validate:"required"`
}

func (h *Handlers) GetUser(c echo.Context) error {
	name, err := h.dbProvider.SelectUser()
	if err != nil {
		c.Logger().Error("Ошибка получения данных из БД!", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения данных из БД!"})
	}
	return c.JSON(http.StatusOK, map[string]string{"msg": "hello, " + name + "!"})
}

func (h *Handlers) PostUser(c echo.Context) error {
	name := c.FormValue("name")
	if name == "" {
		c.Logger().Error("Пользователь не задан!")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Пользователь не задан!"})
	}
	err := h.dbProvider.InsertHello(name)
	if err != nil {
		c.Logger().Error("Ошибка вставки пользователя в БД!")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка вставки пользователя в БД!"})
	}
	return c.JSON(http.StatusOK, map[string]string{"msg": name})
}

func (dp *DatabaseProvider) SelectUser() (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT message FROM query ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO query(message) VALUES ($1)", msg)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер для БД с набором методов
	dp := DatabaseProvider{db: db}
	// Создаем экземпляр структуры с набором обработчиков
	h := Handlers{dbProvider: dp}

	e := echo.New()
	e.GET("/api/user", h.GetUser)
	e.POST("/api/user", h.PostUser)
	e.Logger.Fatal(e.Start(":9000"))
}
