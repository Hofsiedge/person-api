package nationalize

import (
	"fmt"
	"net/http"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/filler"
)

// nationalizer

type Nationalizer filler.Filler[domain.Nationality, NationalizerValidResponse]

type CountryData struct {
	CountryID string `json:"country_id"` //nolint:tagliatelle
}

type NationalizerValidResponse struct {
	Country []CountryData `json:"country"`
}

func (nvr NationalizerValidResponse) Convert() (domain.Nationality, error) {
	if len(nvr.Country) == 0 {
		return "", filler.ErrNotFound
	}

	nationality := domain.Nationality(nvr.Country[0].CountryID)
	if !nationality.Valid() {
		return "", fmt.Errorf("%w: invalid value for nationality: %v", filler.ErrConversion, nationality)
	}

	return nationality, nil
}

func New(baseURL string, token *string, client *http.Client) Nationalizer {
	return Nationalizer(filler.New[domain.Nationality, NationalizerValidResponse](baseURL, token, client))
}
