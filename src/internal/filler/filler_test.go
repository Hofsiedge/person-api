package filler_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Hofsiedge/person-api/internal/filler"
	testutils "github.com/Hofsiedge/person-api/internal/filler/test_utils"
)

type resp struct {
	Qux *string `json:"qux"`
}

func (r resp) Convert() (string, error) {
	if r.Qux == nil {
		return "", filler.ErrNotFound
	}

	if len(*r.Qux) < 3 {
		return "", filler.ErrConversion
	}

	return *r.Qux, nil
}

//nolint:funlen
func TestFiller(t *testing.T) {
	t.Parallel()

	testCases := []testutils.TestCase[string, resp]{
		{
			CaseName: "valid",
			Headers: testutils.Headers{
				Limit:     &testutils.Wrapper[string]{Value: "1000"},
				Remaining: &testutils.Wrapper[string]{Value: "123"},
				Reset:     &testutils.Wrapper[string]{Value: "15000"},
			},
			Name:       "Dmitriy",
			Token:      nil,
			Body:       []byte(`{"qux":"res_string","buzz":23}`),
			StatusCode: http.StatusOK,
			Err:        nil,
			Result:     "res_string",
			Fields: &testutils.Fields{
				Limit:     1000,
				Remaining: 123,
				Reset:     time.Now().Add(15000 * time.Second),
			},
		},
		{
			CaseName: "not found",
			Headers: testutils.Headers{
				Limit:     &testutils.Wrapper[string]{Value: "100"},
				Remaining: &testutils.Wrapper[string]{Value: "50"},
				Reset:     &testutils.Wrapper[string]{Value: "1000"},
			},
			Name:       "naaaame",
			Token:      nil,
			Body:       []byte(`{"qux":null,"buzz":23}`),
			StatusCode: http.StatusOK,
			Err:        filler.ErrNotFound,
			Result:     "",
			Fields: &testutils.Fields{
				Limit:     100,
				Remaining: 50,
				Reset:     time.Now().Add(1000 * time.Second),
			},
		},
		{
			CaseName: "reached limit",
			Headers: testutils.Headers{
				Limit:     &testutils.Wrapper[string]{Value: "100"},
				Remaining: &testutils.Wrapper[string]{Value: "0"},
				Reset:     &testutils.Wrapper[string]{Value: "1000"},
			},
			Name:       "Bill",
			Token:      nil,
			Body:       []byte(`{"error":"Request limit reached"}`),
			StatusCode: http.StatusTooManyRequests,
			Err:        filler.ErrLimitReached,
			Result:     "",
			Fields: &testutils.Fields{
				Limit:     100,
				Remaining: 0,
				Reset:     time.Now().Add(1000 * time.Second),
			},
		},
	}

	testutils.RunSubtests(t, testCases)
}
