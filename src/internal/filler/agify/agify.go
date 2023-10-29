package agify

import (
	"net/http"

	"github.com/Hofsiedge/person-api/internal/filler"
)

// agifier

type Agifier filler.Filler[int, AgifierValidResponse]

type AgifierValidResponse struct {
	Age *int `json:"age"`
}

func (avr AgifierValidResponse) Convert() (int, error) {
	if avr.Age == nil {
		return 0, filler.ErrNotFound
	}

	return *avr.Age, nil
}

func New(baseURL string, token *string, client *http.Client) Agifier {
	return Agifier(filler.New[int, AgifierValidResponse](baseURL, token, client))
}
