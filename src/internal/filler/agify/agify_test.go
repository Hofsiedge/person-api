package agify_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Hofsiedge/person-api/internal/filler"
	"github.com/Hofsiedge/person-api/internal/filler/agify"
	testutils "github.com/Hofsiedge/person-api/internal/filler/test_utils"
)

func TestAgify(t *testing.T) {
	t.Parallel()

	// these are the only interesting test cases
	// other cases are generic and should be tested in the filler package
	testCases := []testutils.TestCase[int, agify.AgifierValidResponse]{
		{
			CaseName: "valid",
			Headers: testutils.Headers{
				Limit:     &testutils.Wrapper[string]{Value: "1000"},
				Remaining: &testutils.Wrapper[string]{Value: "100"},
				Reset:     &testutils.Wrapper[string]{Value: "15000"},
			},
			Name:       "Michael",
			Token:      nil,
			Body:       []byte(`{"count":298219,"name":"michael","age":62}`),
			StatusCode: http.StatusOK,
			Err:        nil,
			Result:     62,
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
			Name:       "Michaellllll",
			Token:      nil,
			Body:       []byte(`{"count":0,"name":"michaellllll","age":null}`),
			StatusCode: http.StatusOK,
			Err:        filler.ErrNotFound,
			Result:     0,
			Fields: &testutils.Fields{
				Limit:     2000,
				Remaining: 10,
				Reset:     time.Now().Add(1000 * time.Second),
			},
		},
	}

	testutils.RunSubtests(t, testCases)
}
