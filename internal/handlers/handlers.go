package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sol1corejz/enricher/internal/domain/models"
	"github.com/sol1corejz/enricher/internal/services/enricher"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

// DataWithFilters godoc
// @Summary Get filtered users
// @Description Retrieve users with optional filters
// @Tags users
// @Accept json
// @Produce json
// @Param name query string false "Filter by name (partial match)"
// @Param surname query string false "Filter by surname (partial match)"
// @Param patronymic query string false "Filter by patronymic (partial match)"
// @Param ageFrom query int false "Minimum age"
// @Param ageTo query int false "Maximum age"
// @Param sex query string false "Filter by sex (male/female)"
// @Param country query string false "Filter by country code (partial match)"
// @Param limit query int false "Pagination limit (default 10)"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router / [get]
func DataWithFilters(ctx *fiber.Ctx) error {
	service, ok := ctx.Locals("enricherService").(*enricher.Enricher)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "service not available",
		})
	}

	// Создаем фильтр из query-параметров
	filter := models.UserFilter{
		Name:       ctx.Query("name"),
		Surname:    ctx.Query("surname"),
		Patronymic: ctx.Query("patronymic"),
		Sex:        ctx.Query("sex"),
		Country:    ctx.Query("country"),
	}

	// Обрабатываем числовые параметры
	if ageFrom := ctx.Query("ageFrom"); ageFrom != "" {
		if age, err := strconv.Atoi(ageFrom); err == nil {
			filter.AgeFrom = age
		}
	}

	if ageTo := ctx.Query("ageTo"); ageTo != "" {
		if age, err := strconv.Atoi(ageTo); err == nil {
			filter.AgeTo = age
		}
	}

	if filter.Sex != "" && filter.Sex != "male" && filter.Sex != "female" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid sex value, must be 'male' or 'female'",
		})
	}

	if limit := ctx.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := ctx.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	// Получаем данные с фильтрами
	users, err := service.GetUsers(ctx.Context(), filter)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to get users",
			"details": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"count": len(users),
		"users": users,
	})
}

// Delete godoc
// @Summary Delete a user
// @Description Delete user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.DeleteUserPayload true "Delete request"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /delete [post]
func Delete(ctx *fiber.Ctx) error {
	service, ok := ctx.Locals("enricherService").(*enricher.Enricher)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "service not available",
		})
	}

	var payloadData models.DeleteUserPayload

	err := ctx.BodyParser(&payloadData)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	err = service.DeleteUser(context.Background(), payloadData.ID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user deleted",
	})
}

// Edit godoc
// @Summary Update a user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.EditUserPayload true "Update data"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /edit [post]
func Edit(ctx *fiber.Ctx) error {
	service, ok := ctx.Locals("enricherService").(*enricher.Enricher)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "service not available",
		})
	}

	var payloadData models.EditUserPayload
	if err := ctx.BodyParser(&payloadData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request payload",
			"details": err.Error(),
		})
	}

	// Получаем текущие данные пользователя
	existingUser, err := service.GetUser(ctx.Context(), payloadData.ID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "user not found",
			"details": err.Error(),
		})
	}

	// Обновляем только те поля, которые пришли в запросе
	if payloadData.Name != "" {
		existingUser.Name = payloadData.Name
	}
	if payloadData.Surname != "" {
		existingUser.Surname = payloadData.Surname
	}
	if payloadData.Patronymic != "" {
		existingUser.Patronymic = payloadData.Patronymic
	}

	// Сохраняем обновленные данные
	updatedUser, err := service.EditUser(ctx.Context(), existingUser)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to update user",
			"details": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "user updated successfully",
		"user":    updatedUser,
	})
}

// Add godoc
// @Summary Create a new user
// @Description Add new user with data enrichment
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.SaveUserPayload true "User data"
// @Success 201 {object} map[string]interface{} "User created"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 422 {object} map[string]interface{} "Validation error"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /add [post]
func Add(ctx *fiber.Ctx) error {
	service, ok := ctx.Locals("enricherService").(*enricher.Enricher)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "service not available",
		})
	}

	var payloadData models.SaveUserPayload
	if err := ctx.BodyParser(&payloadData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request payload",
			"details": err.Error(),
		})
	}

	// Validate all name fields
	if err := ValidateAllNames(payloadData.Name, payloadData.Surname, payloadData.Patronymic); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid name format",
			"details": err.Error(),
		})
	}

	// Обогащаем данные
	enrichedUser, err := enrichUserData(payloadData)
	if err != nil {
		return ctx.Status(fiber.StatusFailedDependency).JSON(fiber.Map{
			"error":   "failed to enrich user data",
			"details": err.Error(),
		})
	}

	// Сохраняем в базу
	id, err := service.SaveUser(ctx.Context(), enrichedUser)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to save user",
			"details": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
	})
}

func enrichUserData(userData models.SaveUserPayload) (models.EnrichedUser, error) {
	var enriched models.EnrichedUser
	enriched.Name = userData.Name
	enriched.Surname = userData.Surname
	enriched.Patronymic = userData.Patronymic

	// Получаем возраст
	age, err := getAge(userData.Name)
	if err != nil {
		return models.EnrichedUser{}, fmt.Errorf("failed to get age: %w", err)
	}
	enriched.Age = age

	// Получаем пол
	sex, err := getGender(userData.Name)
	if err != nil {
		return models.EnrichedUser{}, fmt.Errorf("failed to get gender: %w", err)
	}
	enriched.Sex = sex

	// Получаем национальность
	countries, err := getNationality(userData.Name)
	if err != nil {
		return models.EnrichedUser{}, fmt.Errorf("failed to get nationality: %w", err)
	}
	enriched.Country = countries

	return enriched, nil
}

// Вспомогательные функции для запросов к API

func getAge(name string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.agify.io/?name=%s", name))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Age int `json:"age"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Age, nil
}

func getGender(name string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.genderize.io/?name=%s", name))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Gender string `json:"gender"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Gender, nil
}

func getNationality(name string) ([]models.Country, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.nationalize.io/?name=%s", name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	countries := make([]models.Country, len(result.Country))
	for i, c := range result.Country {
		countries[i] = models.Country{
			CountryID:   c.CountryID,
			Probability: c.Probability,
		}
	}

	return countries, nil
}

// validateNameField validates a single name field (first name, last name or patronymic)
func validateNameField(fieldName, value string, required bool) error {
	// Check if required field is empty
	if required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	// Skip validation if field is not required and empty
	if !required && strings.TrimSpace(value) == "" {
		return nil
	}

	// Check length
	runes := []rune(value)
	if len(runes) > 100 {
		return fmt.Errorf("%s is too long (max 100 characters)", fieldName)
	}

	// Check each character
	for i, r := range value {
		switch {
		case unicode.IsLetter(r):
			continue
		case r == ' ' || r == '-' || r == '\'':
			// Check for leading/trailing special characters
			if i == 0 || i == len(runes)-1 {
				return fmt.Errorf("%s cannot start or end with special characters", fieldName)
			}
			// Check for consecutive special characters
			if i > 0 && (r == rune(value[i-1])) {
				return fmt.Errorf("%s cannot have consecutive special characters", fieldName)
			}
		default:
			return fmt.Errorf("%s contains invalid characters - only letters, spaces, hyphens and apostrophes are allowed", fieldName)
		}
	}

	return nil
}

// ValidateAllNames validates all name fields at once
func ValidateAllNames(firstName, lastName, patronymic string) error {
	if err := validateNameField("first name", firstName, true); err != nil {
		return err
	}
	if err := validateNameField("last name", lastName, true); err != nil {
		return err
	}
	if err := validateNameField("patronymic", patronymic, false); err != nil {
		return err
	}
	return nil
}
