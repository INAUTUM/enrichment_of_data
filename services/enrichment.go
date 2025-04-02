package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"project/models"
)

// EnrichPerson получает доп. данные по API и обновляет `person`
func EnrichPerson(person *models.Person) error {
	ageURL := fmt.Sprintf("https://api.agify.io/?name=%s", person.Name)
	genderURL := fmt.Sprintf("https://api.genderize.io/?name=%s", person.Name)
	nationalityURL := fmt.Sprintf("https://api.nationalize.io/?name=%s", person.Name)

	// Функция запроса к API
	fetchData := func(url string, target interface{}) error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return json.NewDecoder(resp.Body).Decode(target)
	}

	var ageResponse struct{ Age int }
	var genderResponse struct{ Gender string }
	var nationalityResponse struct {
		Country []struct {
			CountryID string `json:"country_id"`
		} `json:"country"`
	}

	if err := fetchData(ageURL, &ageResponse); err != nil {
		return err
	}
	if err := fetchData(genderURL, &genderResponse); err != nil {
		return err
	}
	if err := fetchData(nationalityURL, &nationalityResponse); err != nil {
		return err
	}

	person.Age = ageResponse.Age
	person.Gender = genderResponse.Gender
	if len(nationalityResponse.Country) > 0 {
		person.Nationality = nationalityResponse.Country[0].CountryID
	}

	return nil
}