package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// @title People API
// @version 1.0
// @description API для работы с людьми, включая создание, обновление, удаление и получение данных.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @host localhost:8081
// @BasePath /

type Person struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic,omitempty"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
}

var db *gorm.DB

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[WARN] .env файл не найден, используем стандартные настройки")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://admin:admin@localhost:5432/test?sslmode=disable"
	}
	fmt.Println("[INFO] Подключение к БД:", dsn)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("[ERROR] Ошибка подключения к БД:", err)
	}
	db.AutoMigrate(&Person{})
}

// fetchData получает возраст, пол и национальность
func fetchData(name string) (int, string, string, error) {
	age, err := fetchAge(name)
	if err != nil {
		return 0, "", "", err
	}
	gender, err := fetchGender(name)
	if err != nil {
		return 0, "", "", err
	}
	nationality, err := fetchNationality(name)
	if err != nil {
		return 0, "", "", err
	}
	return age, gender, nationality, nil
}

// fetchAge запрашивает возраст с agify.io
func fetchAge(name string) (int, error) {
	url := fmt.Sprintf("https://api.agify.io/?name=%s", name)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data struct {
		Age int `json:"age"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	return data.Age, nil
}

// fetchGender запрашивает пол с genderize.io
func fetchGender(name string) (string, error) {
	url := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Gender string `json:"gender"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Gender, nil
}

// fetchNationality запрашивает национальность с nationalize.io
func fetchNationality(name string) (string, error) {
	url := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Country []struct {
			CountryID string `json:"country_id"`
		} `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Country) > 0 {
		return data.Country[0].CountryID, nil
	}
	return "Unknown", nil
}

// CreatePerson godoc
// @Summary Создание нового человека
// @Description Принимает ФИО, обогащает данными (возраст, пол, национальность) и сохраняет в БД
// @Tags People
// @Accept  json
// @Produce  json
// @Param person body models.Person true "Имя, фамилия и отчество"
// @Success 201 {object} models.Person
// @Failure 400 {object} object "{\"error": \"Invalid input"}"
// @Failure 500 {object} object "{\"error": \"Failed to save person"}"
// @Router /people [post]
func createPerson(c *gin.Context) {
	var person Person
	if err := c.ShouldBindJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	age, gender, nationality, err := fetchData(person.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обогатить данные"})
		return
	}

	person.Age = age
	person.Gender = gender
	person.Nationality = nationality
	db.Create(&person)
	c.JSON(http.StatusCreated, person)
}

// getPeople godoc
// @Summary Получение списка людей с фильтрацией и пагинацией
// @Description Возвращает список людей с возможностью фильтрации по имени и полу, а также пагинации
// @Tags People
// @Accept  json
// @Produce  json
// @Param name query string false "Имя"
// @Param gender query string false "Пол"
// @Param limit query int false "Лимит"
// @Param offset query int false "Смещение"
// @Success 200 {array} models.Person
// @Failure 400 {object} object "{\"error": \"Invalid query parameters"}"
// @Router /people [get]
func getPeople(c *gin.Context) {
	var people []Person
	query := db

	// Фильтрация
	if name := c.Query("name"); name != "" {
		query = query.Where("name = ?", name)
	}
	if gender := c.Query("gender"); gender != "" {
		query = query.Where("gender = ?", gender)
	}

	// Пагинация
	limit := 10
	offset := 0
	if c.Query("limit") != "" {
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
	}
	if c.Query("offset") != "" {
		fmt.Sscanf(c.Query("offset"), "%d", &offset)
	}

	query.Limit(limit).Offset(offset).Find(&people)
	c.JSON(http.StatusOK, people)
}

// updatePerson обновляет информацию о человеке
// UpdatePerson godoc
// @Summary Обновление информации о человеке
// @Description Обновляет информацию о человеке по его ID
// @Tags People
// @Accept  json
// @Produce  json
// @Param id path int true "ID человека"
// @Param person body models.Person true "Данные для обновления"
// @Success 200 {object} models.Person
// @Failure 400 {object} object "{\"error": \"Invalid input"}"
// @Failure 404 {object} object "{\"error": \"Person not found\"}"
// @Router /people/{id} [put]
func updatePerson(c *gin.Context) {
	id := c.Param("id")
	var person Person
	if err := db.First(&person, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	if err := c.ShouldBindJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&person)
	c.JSON(http.StatusOK, person)
}

// deletePerson удаляет человека по ID
// DeletePerson godoc
// @Summary Удаление человека по ID
// @Description Удаляет человека из базы данных по его ID
// @Tags People
// @Accept  json
// @Produce  json
// @Param id path int true "ID человека"
// @Success 200 {object} object "{\"message\": \"Пользователь удален\"}"
// @Failure 400 {object} object "{\"error\": \"Invalid ID"}"
// @Failure 404 {object} object "{\"error\": \"Person not found\"}"
// @Failure 500 {object} object "{\"error\": \"Failed to delete person\"}"
// @Router /people/{id} [delete]
func deletePerson(c *gin.Context) {
	id := c.Param("id")
	if err := db.Delete(&Person{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удален"})
}

// setupRouter настраивает маршруты
func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/people", createPerson)
	r.GET("/people", getPeople)
	r.PUT("/people/:id", updatePerson)
	r.DELETE("/people/:id", deletePerson)

	return r
}

func main() {
	r := setupRouter()

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8081")
}