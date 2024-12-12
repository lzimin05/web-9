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

func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectHello()
	if err != nil {
		c.Logger().Error("Ошибка при получении сообщения из БД:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получения данных!"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": msg})
}

func (h *Handlers) PostHello(c echo.Context) error {
	var input CountRequest
	if err := c.Bind(&input); err != nil {
		c.Logger().Error("Ошибка парсинга JSON!:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинга JSON"})
	}
	err := h.dbProvider.InsertHello(input.Msg)
	if err != nil {
		c.Logger().Error("Ошибка при вставке сообщения в БД:")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавления записи"})
	}
	return c.JSON(http.StatusOK, input.Msg)
}

func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (dp *DatabaseProvider) InsertHello(msg string) error {
	if msg == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "request body is empty or msg JSON is an empty string!")
	}
	_, err := dp.db.Exec("INSERT INTO hello(message) VALUES ($1)", msg)
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
	e.GET("/get", h.GetHello)
	e.POST("/post", h.PostHello)
	e.Logger.Fatal(e.Start(":8081"))
}
