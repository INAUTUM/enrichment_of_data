package handlers

import (
	"log"
	"net/http"
	"project/models"
	"project/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
func CreatePerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var person models.Person
		if err := c.ShouldBindJSON(&person); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Обогащаем данные перед сохранением
		if err := services.EnrichPerson(&person); err != nil {
			log.Printf("[ERROR] Ошибка при обогащении: %v", err)
		}

		// Сохраняем в БД
		if err := db.Create(&person).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save person"})
			return
		}

		c.JSON(http.StatusCreated, person)
	}
}