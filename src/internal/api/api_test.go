package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Hofsiedge/person-api/internal/api"
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/repo"
	"github.com/Hofsiedge/person-api/internal/repo/mock"
	"github.com/Hofsiedge/person-api/internal/utils"
	"github.com/google/uuid"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

type testCase struct {
	init   func(t *testing.T, people repo.PersonRepo) (req *http.Request, check func(response *http.Response))
	name   string
	status int
}

func subtests(t *testing.T, testCases []testCase) {
	t.Helper()

	for _, tCase := range testCases {
		test := tCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			people := mock.New()
			request, check := test.init(t, people) //nolint:bodyclose
			result := serve(t, request, people)
			defer result.Body.Close()

			// check status codes
			if test.status != result.StatusCode {
				t.Errorf("unexpected status code: expected %d, got %d",
					test.status, result.StatusCode)
			}
			// run checks
			if check != nil {
				check(result)
			}
		})
	}
}

func checkNoBody(t *testing.T, response *http.Response) {
	t.Helper()

	if response.Body == nil {
		return
	}

	var data map[string]any

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	if len(bytes) == 0 {
		return
	}

	if err = json.Unmarshal(bytes, &data); err != nil {
		t.Fatalf("could not unmarshal unexpected response body: %v\nbody: %s", err, bytes)
	}

	t.Errorf("unexpected response body: %v", data)
}

func unmarshalJSONBody[T any](t *testing.T, response *http.Response) T {
	t.Helper()

	var body T

	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error reading response body: %v", err)
	}

	if err = json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unexpected error unmarshalling JSON response: %v\ndata: %s", err, data)
	}

	return body
}

func checkStringBody(t *testing.T, response *http.Response, pattern *regexp.Regexp) {
	t.Helper()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error reading response body: %v", err)
	}

	if !pattern.Match(data) {
		t.Errorf("response body does not match the pattern.\nresponse: %s", data)
	}
}

func checkBody[T any](t *testing.T, response *http.Response, expected T) {
	t.Helper()

	body := unmarshalJSONBody[T](t, response)
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("body mismatch: expected %v, got %v", expected, body)
	}
}

// initialize a server and run request against it
func serve(t *testing.T, request *http.Request, people repo.PersonRepo) *http.Response {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		AddSource:   false,
		Level:       nil,
		ReplaceAttr: nil,
	}))

	server, err := api.New(people, logger)
	if err != nil {
		t.Fatalf("error creating a server: %v", err)
	}
	// validator
	spec, err := api.GetSwagger()
	if err != nil {
		t.Fatalf("error loading openapi spec: %s", err.Error())
	}
	// do not validate server names
	spec.Servers = nil
	oapiValidator := middleware.OapiRequestValidator(spec)

	handler := api.HandlerWithOptions(
		api.NewStrictHandler(server, []api.StrictMiddlewareFunc{}),
		api.GorillaServerOptions{ //nolint:exhaustruct
			Middlewares: []api.MiddlewareFunc{oapiValidator},
		},
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	result := recorder.Result()

	return result
}

func TestGet(t *testing.T) {
	t.Parallel()

	makeGetRequest := func(id any) *http.Request {
		return httptest.NewRequest(http.MethodGet, fmt.Sprintf("/person/%s", id), nil)
	}

	testCases := []testCase{
		{
			name: "not found",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				request := makeGetRequest(uuid.New())

				return request, func(response *http.Response) {
					checkNoBody(t, response)
				}
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				request := makeGetRequest("124390845")

				return request, func(response *http.Response) {
					pattern := regexp.MustCompile("Invalid format for parameter personID")
					checkStringBody(t, response, pattern)
				}
			},
			status: http.StatusBadRequest,
		},
		{
			name: "found",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				person := utils.MakePerson()
				personID, err := people.Create(context.Background(), person)
				if err != nil {
					t.Fatalf("error initializing repo: %v", err)
				}
				person.ID = personID
				request := makeGetRequest(personID)
				body := api.PersonGet200JSONResponse{
					Age:         person.Age,
					Id:          person.ID,
					Name:        person.Name,
					Nationality: person.Nationality,
					Patronymic:  person.Patronymic,
					Sex:         api.Sex(person.Sex),
					Surname:     person.Surname,
				}

				return request, func(response *http.Response) {
					checkBody(t, response, body)
				}
			},
			status: http.StatusOK,
		},
	}

	subtests(t, testCases)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	makeDeleteRequest := func(id any) *http.Request {
		return httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/person/%s", id), nil)
	}

	testCases := []testCase{
		{
			name: "not found",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				return makeDeleteRequest(uuid.New()), nil
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				request := makeDeleteRequest("20934822")

				return request, func(response *http.Response) {
					pattern := regexp.MustCompile("Invalid format for parameter personID")
					checkStringBody(t, response, pattern)
				}
			},
			status: http.StatusBadRequest,
		},
		{
			name: "valid",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				person := utils.MakePerson()
				personID, err := people.Create(context.Background(), person)
				if err != nil {
					t.Fatalf("error initializing repo: %v", err)
				}

				return makeDeleteRequest(personID), func(response *http.Response) {
					checkNoBody(t, response)
					_, err := people.GetByID(context.Background(), personID)
					if !errors.Is(err, repo.ErrNotFound) {
						t.Errorf("deleted person is still in the repo")
					}
				}
			},
			status: http.StatusOK,
		},
	}

	subtests(t, testCases)
}

//nolint:funlen
func TestPut(t *testing.T) {
	t.Parallel()

	makePutRequest := func(personID any, body *api.PersonPutJSONRequestBody) *http.Request {
		var reader io.Reader

		if body != nil {
			data, err := json.Marshal(*body)
			if err != nil {
				t.Fatalf("could not marshal put body: %v", err)
			}

			reader = bytes.NewReader(data)
		}

		request := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/person/%s", personID),
			reader,
		)
		request.Header.Set("Content-Type", "application/json")

		return request
	}

	makeBody := func() api.PersonPutJSONRequestBody {
		person := utils.MakePerson()
		body := api.PersonPutJSONRequestBody{
			Age:         person.Age,
			Name:        person.Name,
			Nationality: person.Nationality,
			Patronymic:  person.Patronymic,
			Sex:         api.Sex(person.Sex),
			Surname:     person.Surname,
		}

		return body
	}

	testCases := []testCase{
		{
			name: "not found",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				body := makeBody()

				return makePutRequest(uuid.New(), &body), func(response *http.Response) {
					checkNoBody(t, response)
				}
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				body := makeBody()
				request := makePutRequest("20934822", &body)

				return request, func(response *http.Response) {
					pattern := regexp.MustCompile("Invalid format for parameter personID")
					checkStringBody(t, response, pattern)
				}
			},
			status: http.StatusBadRequest,
		},
		{
			name: "valid",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				person := utils.MakePerson()
				personID, err := people.Create(context.Background(), person)
				if err != nil {
					t.Fatalf("error initializing repo: %v", err)
				}
				newPerson := utils.MakePerson()
				request := makePutRequest(personID, &api.PersonPutJSONRequestBody{
					Age:         newPerson.Age,
					Name:        newPerson.Name,
					Nationality: newPerson.Nationality,
					Patronymic:  newPerson.Patronymic,
					Sex:         api.Sex(newPerson.Sex),
					Surname:     newPerson.Surname,
				})

				return request, func(response *http.Response) {
					checkNoBody(t, response)

					personAfter, err := people.GetByID(context.Background(), personID)
					if err != nil {
						t.Fatalf("could not get new value: %v", err)
					}
					newPerson.ID = personID
					if !reflect.DeepEqual(newPerson, personAfter) {
						t.Errorf("replaced person does not match the provided value: expected %v, got %v",
							newPerson, personAfter)
					}
				}
			},
			status: http.StatusOK,
		},
	}

	subtests(t, testCases)
}

//nolint:funlen
func TestPost(t *testing.T) {
	t.Parallel()

	makePostRequest := func(body any) *http.Request {
		var reader io.Reader

		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("could not marshal put body: %v", err)
			}

			reader = bytes.NewReader(data)
		}

		request := httptest.NewRequest(
			http.MethodPost,
			"/person",
			reader,
		)
		request.Header.Set("Content-Type", "application/json")

		return request
	}

	testCases := []testCase{
		{
			name: "valid",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				person := utils.MakePerson()
				request := makePostRequest(api.PersonPostJSONRequestBody{
					Name:       person.Name,
					Patronymic: person.Patronymic,
					Surname:    person.Surname,
				})

				return request, func(response *http.Response) {
					personID := unmarshalJSONBody[api.PersonPost201JSONResponse](t, response)

					personAfter, err := people.GetByID(context.Background(), uuid.UUID(personID))
					if err != nil {
						t.Fatalf("Person was not saved after response")
					}

					afterData := []string{personAfter.Name, personAfter.Patronymic, personAfter.Surname}
					beforeData := []string{person.Name, person.Patronymic, person.Surname}
					if !slices.Equal(afterData, beforeData) {
						t.Errorf("incorrect data for [Name, Patronymic, Surname]: expected %v, got %v",
							beforeData, afterData)
					}
				}
			},
			status: http.StatusCreated,
		},
		{
			name: "invalid body",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				requestBody := `{"name":"Quux","surname":"Buzz"}`
				request := httptest.NewRequest(
					http.MethodPost,
					"/person",
					strings.NewReader(requestBody))
				request.Header.Set("Content-Type", "application/json")

				return request, func(response *http.Response) {
					checkStringBody(t, response, regexp.MustCompile(`property "patronymic" is missing`))
				}
			},
			status: http.StatusBadRequest,
		},
	}

	subtests(t, testCases)
}

//nolint:funlen
func TestPatch(t *testing.T) {
	t.Parallel()

	makePatchRequest := func(personID any, body any) *http.Request {
		var reader io.Reader

		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("could not marshal put body: %v", err)
			}

			reader = bytes.NewReader(data)
		}

		request := httptest.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/person/%s", personID),
			reader,
		)
		request.Header.Set("Content-Type", "application/json")

		return request
	}

	testCases := []testCase{
		{
			name: "not found",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				newAge := 60

				return makePatchRequest(uuid.New(), api.PersonPatchJSONRequestBody{ //nolint:exhaustruct
						Age: &newAge,
					}), func(response *http.Response) {
						checkNoBody(t, response)
					}
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, _ repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				newAge := 60

				return makePatchRequest("quux", api.PersonPatchJSONRequestBody{ //nolint:exhaustruct
						Age: &newAge,
					}), func(response *http.Response) {
						checkStringBody(t, response, regexp.MustCompile(`Invalid format for parameter personID`))
					}
			},
			status: http.StatusBadRequest,
		},
		{
			name: "valid age update",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				person := utils.MakePerson()
				personID, err := people.Create(context.Background(), person)
				if err != nil {
					t.Fatalf("error initializing repo: %v", err)
				}
				newAge := person.Age / 2
				request := makePatchRequest(personID, api.PersonPatchJSONRequestBody{ //nolint:exhaustruct
					Age: &newAge,
				})

				return request, func(response *http.Response) {
					checkNoBody(t, response)

					personAfter, err := people.GetByID(context.Background(), personID)
					if err != nil {
						t.Fatalf("Person was deleted after response")
					}

					if personAfter.Age != newAge {
						t.Error("the age did not change")
					}

					beforeData := []any{
						person.Name, person.Surname, person.Patronymic,
						person.Nationality, person.Sex,
					}
					afterData := []any{
						personAfter.Name, personAfter.Surname, personAfter.Patronymic,
						personAfter.Nationality, personAfter.Sex,
					}

					if !slices.Equal(afterData, beforeData) {
						t.Errorf("other fields were updated: expected %v, got %v",
							beforeData, afterData)
					}
				}
			},
			status: http.StatusOK,
		},
	}

	subtests(t, testCases)
}

func makeListRequest(personFilter domain.PersonFilter, paginationFilter *domain.PaginationFilter) *http.Request {
	values := url.Values{}
	// personFilter values
	if personFilter.Sex != nil {
		values.Add("sex", string(*personFilter.Sex))
	}

	if personFilter.AgeMin != nil {
		values.Add("age_min", strconv.Itoa(*personFilter.AgeMin))
	}

	if personFilter.AgeMax != nil {
		values.Add("age_max", strconv.Itoa(*personFilter.AgeMax))
	}

	if personFilter.Nationality != nil {
		values.Add("nationality", *personFilter.Nationality)
	}

	if personFilter.Name != nil {
		values.Add("name", *personFilter.Name)
	}

	if personFilter.Patronymic != nil {
		values.Add("patronymic", *personFilter.Patronymic)
	}

	if personFilter.Surname != nil {
		values.Add("surname", *personFilter.Surname)
	}
	// paginationFilter values
	if paginationFilter != nil {
		values.Add("offset", strconv.Itoa(paginationFilter.Offset))
		values.Add("limit", strconv.Itoa(paginationFilter.Limit))
	}

	path := "/person"
	if q := values.Encode(); len(q) != 0 {
		path += "?" + q
	}

	return httptest.NewRequest(http.MethodGet, path, nil)
}

//nolint:funlen
func TestList(t *testing.T) {
	t.Parallel()

	fillDB := func(people repo.PersonRepo) {
		for i := 0; i < 100; i++ {
			person := utils.MakePerson()
			if _, err := people.Create(context.Background(), person); err != nil {
				t.Fatalf("error initializing repo: %v", err)
			}
		}
	}

	testCases := []testCase{
		{
			name: "valid",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) { //nolint:thelper
				minAge := 20
				nameFragment := "a"
				personFilter := domain.PersonFilter{ //nolint:exhaustruct
					AgeMin: &minAge,
					Name:   &nameFragment,
				}
				paginationFilter := domain.PaginationFilter{Limit: 2, Offset: 1}

				var expected domain.Page[domain.Person]
				// fill DB until there is enough to fill a page
				for {
					fillDB(people)
					var err error
					expected, err = people.List(context.Background(), personFilter, paginationFilter)
					if err != nil {
						t.Fatalf("error initializing repo: %v", err)
					}
					if len(expected.Items) == paginationFilter.Limit {
						break
					}
				}

				request := makeListRequest(personFilter, &paginationFilter)

				return request, func(response *http.Response) {
					records := make([]api.PersonFullWithID, expected.CurrentLimit)
					for i, person := range expected.Items {
						records[i] = api.PersonFullWithID{
							Age:         person.Age,
							Id:          person.ID,
							Name:        person.Name,
							Nationality: person.Nationality,
							Patronymic:  person.Patronymic,
							Sex:         api.Sex(person.Sex),
							Surname:     person.Surname,
						}
					}
					result := api.PersonList200JSONResponse{
						Pagination: api.PaginationOffsetLimit{
							CurrentLimit:  expected.CurrentLimit,
							CurrentOffset: expected.CurrentOffset,
							TotalItems:    expected.TotalItems,
						},
						People: records,
					}
					checkBody(t, response, result)
				}
			},
			status: http.StatusOK,
		},
	}

	subtests(t, testCases)
}
