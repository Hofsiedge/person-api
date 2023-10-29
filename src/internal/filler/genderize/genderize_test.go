package genderize_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/filler"
	"github.com/Hofsiedge/person-api/internal/filler/genderize"
	testutils "github.com/Hofsiedge/person-api/internal/filler/test_utils"
)

func TestGenderize(t *testing.T) {
	t.Parallel()

	// these are the only interesting test cases
	// other cases are generic and should be tested in the filler package
	testCases := []testutils.TestCase[domain.Sex, genderize.GenderizerValidResponse]{
		{
			CaseName: "valid",
			Headers: testutils.Headers{
				Limit:     &testutils.Wrapper[string]{Value: "1000"},
				Remaining: &testutils.Wrapper[string]{Value: "100"},
				Reset:     &testutils.Wrapper[string]{Value: "15000"},
			},
			Name:  "Ashley",
			Token: nil,
			Body: []byte(
				`{"count":389780,"name":"Ashley","gender":"female","probability":0.99}`),
			StatusCode: http.StatusOK,
			Err:        nil,
			Result:     domain.Female,
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
			Name:       "Ashleyyyyy",
			Token:      nil,
			Body:       []byte(`{"count":0,"name":"Ashleyyyyy","gender":null,"probability":0.0}`),
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
