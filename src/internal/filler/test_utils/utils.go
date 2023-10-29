package testutils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/Hofsiedge/person-api/internal/filler"
)

const AcceptableTimeDelta = time.Second * 3

// just a helper type - there are no *string, *int, etc literals, but there are *struct literals
type Wrapper[T any] struct {
	Value T
}

type Headers struct {
	Limit     *Wrapper[string]
	Remaining *Wrapper[string]
	Reset     *Wrapper[string]
}

type Fields struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

type TestCase[T any, C filler.Converter[T]] struct {
	CaseName   string
	Headers    Headers
	Name       string
	Token      *string
	Body       []byte
	StatusCode int
	Err        error
	Result     T
	Fields     *Fields
}

func MakeServer[T any, C filler.Converter[T]](t *testing.T, testCase TestCase[T, C]) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(
		http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
			values := url.Values{}

			values.Add("name", testCase.Name)

			if testCase.Token != nil {
				values.Add("apikey", *testCase.Token)
			}

			expectedURL := "/?" + values.Encode()

			if req.URL.String() != expectedURL {
				t.Errorf("URL mismatch: expected %q, got %q", expectedURL, req.URL.String())
			}

			if testCase.Headers.Limit != nil {
				response.Header().Add("x-rate-limit-limit", testCase.Headers.Limit.Value)
			}

			if testCase.Headers.Remaining != nil {
				response.Header().Add("x-rate-limit-remaining", testCase.Headers.Remaining.Value)
			}

			if testCase.Headers.Reset != nil {
				response.Header().Add("x-rate-limit-reset", testCase.Headers.Reset.Value)
			}
			response.WriteHeader(testCase.StatusCode)

			response.Header().Add("content-type", "application/json")

			_, _ = response.Write(testCase.Body)
		}),
	)

	return server
}

//nolint:cyclop
func checkFillerFields[T any, C filler.Converter[T]](t *testing.T, fields *Fields, fill *filler.Filler[T, C]) {
	t.Helper()

	if fields == nil {
		_, errLim := fill.RequestLimit()
		_, errRem := fill.RequestsLeft()
		_, errRes := fill.ResetTime()

		if !(errLim != nil && errRem != nil && errRes != nil) {
			t.Errorf("some fields are accessible when they should not be")
		}

		return
	}

	if limit, err := fill.RequestLimit(); err != nil {
		t.Errorf("could not read request limit: %v", err)
	} else if limit != fields.Limit {
		t.Errorf("limit value mismatch: expected %v, got %v", fields.Limit, limit)
	}

	if left, err := fill.RequestsLeft(); err != nil {
		t.Errorf("could not read requests left: %v", err)
	} else if left != fields.Remaining {
		t.Errorf("requests left value mismatch: expected %v, got %v", fields.Remaining, left)
	}

	if reset, err := fill.ResetTime(); err != nil {
		t.Errorf("could not read request limit: %v", err)
	} else if fields.Reset.Sub(reset).Abs() > AcceptableTimeDelta {
		t.Errorf("reset time value mismatch: expected %v, got %v", fields.Reset, reset)
	}
}

func RunSubtests[T any, C filler.Converter[T]](t *testing.T, testCases []TestCase[T, C]) {
	t.Helper()

	for _, tc := range testCases {
		testCase := tc
		t.Run(testCase.CaseName, func(t *testing.T) {
			t.Parallel()

			server := MakeServer[T, C](t, testCase)
			defer server.Close()

			fill := filler.New[T, C](server.URL, nil, server.Client())

			result, err := fill.Fill(testCase.Name)

			checkFillerFields(t, testCase.Fields, &fill)

			if !errors.Is(err, testCase.Err) {
				t.Fatalf("error mismatch: expected %v, got %v", testCase.Err, err)
			}

			if !reflect.DeepEqual(result, testCase.Result) {
				t.Fatalf("result mismatch: expected %v, got %v", testCase.Result, result)
			}
		})
	}
}
