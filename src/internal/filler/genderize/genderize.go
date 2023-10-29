package genderize

import (
	"fmt"
	"net/http"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/filler"
)

// genderizer

type Genderizer filler.Filler[domain.Sex, GenderizerValidResponse]

type GenderizerValidResponse struct {
	Gender *domain.Sex `json:"gender"`
}

func (gvr GenderizerValidResponse) Convert() (domain.Sex, error) {
	if gvr.Gender == nil {
		return "", filler.ErrNotFound
	}

	if !gvr.Gender.Valid() {
		return "", fmt.Errorf("%w: invalid sex value: %v", filler.ErrConversion, gvr.Gender)
	}

	return *gvr.Gender, nil
}

func New(baseURL string, token *string, client *http.Client) Genderizer {
	return Genderizer(filler.New[domain.Sex, GenderizerValidResponse](baseURL, token, client))
}
