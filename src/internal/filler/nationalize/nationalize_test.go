package nationalize_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/filler"
	"github.com/Hofsiedge/person-api/internal/filler/nationalize"
	testutils "github.com/Hofsiedge/person-api/internal/filler/test_utils"
)

func TestNationalize(t *testing.T) {
	t.Parallel()

	// these are the only interesting test cases
	// other cases are generic and should be tested in the filler package
	testCases := []testutils.TestCase[domain.Nationality, nationalize.NationalizerValidResponse]{
		{
			CaseName: "valid",
			Headers: testutils.Headers{
				Limit:     &testutils.Wrapper[string]{Value: "1000"},
				Remaining: &testutils.Wrapper[string]{Value: "100"},
				Reset:     &testutils.Wrapper[string]{Value: "15000"},
			},
			Name:  "Dmitriy",
			Token: nil,
			Body: []byte(
				`{"count":24968,"name":"Dmitriy","country":` +
					`[{"country_id":"UA","probability":0.419},{"country_id":"RU",` +
					`"probability":0.291},{"country_id":"KZ","probability":0.097},` +
					`{"country_id":"BY","probability":0.069},{"country_id":"IL",` +
					`"probability":0.019}]}`),
			StatusCode: http.StatusOK,
			Err:        nil,
			Result:     "UA",
			Fields: &testutils.Fields{
				Limit:     1000,
				Remaining: 100,
				Reset:     time.Now().Add(15000 * time.Second),
			},
		},
		{
			CaseName: "not found",
			Headers: testutils.Headers{
				Limit:     &testutils.Wrapper[string]{Value: "2000"},
				Remaining: &testutils.Wrapper[string]{Value: "10"},
				Reset:     &testutils.Wrapper[string]{Value: "1000"},
			},
			Name:       "Dmitriyyyyyy",
			Token:      nil,
			Body:       []byte(`{"count":0,"name":"Dmitriyyyyyy","country":[]}`),
			StatusCode: http.StatusOK,
			Err:        filler.ErrNotFound,
			Result:     "",
			Fields: &testutils.Fields{
				Limit:     2000,
				Remaining: 10,
				Reset:     time.Now().Add(1000 * time.Second),
			},
		},
	}

	testutils.RunSubtests(t, testCases)
}
