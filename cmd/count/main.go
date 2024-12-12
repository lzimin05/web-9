package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

func (h *Handlers) GetCount(c echo.Context) error {
	count, err := h.dbProvider.SelectCount()
	if err != nil {
		c.Logger().Error("Ошибка получения данных из БД")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения данных из БД"})
	}
	return c.JSON(http.StatusOK, "count = "+count)
}

func (h *Handlers) PostCount(c echo.Context) error {
	count := c.FormValue("count")
	if count == "" {
		c.Logger().Error("Параметр count не задан")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Параметр count не задан"})
	}
	number, err := strconv.Atoi(count)
	if err != nil {
		c.Logger().Error("count не число!")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "count не число!"})
	}
	err = h.dbProvider.UpdateCount(number)
	if err != nil {
		c.Logger().Error("Ошибка изменения count в БД")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка изменения count в БД"})
	}
	return c.JSON(http.StatusOK, "count успешно изменен")
}

func (dp *DatabaseProvider) SelectCount() (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT count FROM count")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (dp *DatabaseProvider) UpdateCount(count int) error {
	oldcount, err := dp.GetCount()
	if err != nil {
		return err
	}

	_, err = dp.db.Exec("UPDATE count SET count = ($1)", count+oldcount)
	if err != nil {
		return err
	}
	return nil
}

func (dp *DatabaseProvider) GetCount() (int, error) {
	var msg string
	row := dp.db.QueryRow("SELECT count FROM count")
	err := row.Scan(&msg)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(msg)
	if err != nil {
		return 0, err
	}
	return count, nil
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
	e.GET("/count", h.GetCount)
	e.POST("/count", h.PostCount)
	e.Logger.Fatal(e.Start(":3333"))
}
