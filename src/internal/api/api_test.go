package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
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

	"github.com/google/uuid"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type testCase struct {
	name   string
	init   func(t *testing.T, people repo.PersonRepo) (req *http.Request, check func(response *http.Response))
	status int
}

func subtests(t *testing.T, testCases []testCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			people := mock.New()
			request, check := tc.init(t, people)
			result := serve(t, request, people)

			// check status codes
			if tc.status != result.StatusCode {
				t.Errorf("unexpected status code: expected %d, got %d",
					tc.status, result.StatusCode)
			}
			// run checks
			if check != nil {
				check(result)
			}
		})
	}
}

func checkNoBody(t *testing.T, response *http.Response) {
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
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error reading response body: %v", err)
	}
	if !pattern.Match(data) {
		t.Errorf("response body does not match the pattern.\nresponse: %s", data)
	}
}

func checkBody[T any](t *testing.T, response *http.Response, expected T) {
	body := unmarshalJSONBody[T](t, response)
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("body mismatch: expected %v, got %v", expected, body)
	}
}

// initialize a server and run request against it
func serve(t *testing.T, request *http.Request, people repo.PersonRepo) *http.Response {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
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
		api.GorillaServerOptions{
			Middlewares: []api.MiddlewareFunc{
				// TODO: find out why it screws everything over
				oapiValidator,
			},
		},
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	result := recorder.Result()
	return result
}

func generateRandomString(minLength, maxLength uint) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	var builder strings.Builder

	length := minLength + (uint(rand.Int()) % (maxLength + 1 - minLength))
	for i := 0; i < int(length); i++ {
		letter := letters[rand.Int()%len(letters)]
		builder.WriteByte(letter)
	}
	return builder.String()
}

func makePerson() domain.Person {
	capitalizer := cases.Title(language.Und)
	sex := domain.Male
	if rand.Float32() < 0.5 {
		sex = domain.Female
	}
	return domain.Person{
		Name:        capitalizer.String(generateRandomString(2, 10)),
		Surname:     capitalizer.String(generateRandomString(2, 20)),
		Patronymic:  capitalizer.String(generateRandomString(0, 10)),
		Nationality: strings.ToTitle(generateRandomString(2, 2)),
		Sex:         sex,
		Age:         rand.Int() % 120,
		ID:          uuid.New(),
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	makeGetRequest := func(id any) *http.Request {
		return httptest.NewRequest(http.MethodGet, fmt.Sprintf("/person/%s", id), nil)
	}

	testCases := []testCase{
		{
			name: "not found",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				request := makeGetRequest(uuid.New())
				return request, func(response *http.Response) {
					checkNoBody(t, response)
				}
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
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
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				person := makePerson()
				personID, err := people.Create(context.Background(), &person)
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
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				return makeDeleteRequest(uuid.New()), nil
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
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
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				person := makePerson()
				personID, err := people.Create(context.Background(), &person)
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

func TestPut(t *testing.T) {
	t.Parallel()

	makePutRequest := func(id any, body *api.PersonPutJSONRequestBody) *http.Request {
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
			fmt.Sprintf("/person/%s", id),
			reader,
		)
		request.Header.Set("Content-Type", "application/json")
		return request
	}

	makeBody := func() api.PersonPutJSONRequestBody {
		person := makePerson()
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
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				body := makeBody()
				return makePutRequest(uuid.New(), &body), func(response *http.Response) {
					checkNoBody(t, response)
				}
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
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
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				person := makePerson()
				personID, err := people.Create(context.Background(), &person)
				if err != nil {
					t.Fatalf("error initializing repo: %v", err)
				}
				newPerson := makePerson()
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
					if !reflect.DeepEqual(newPerson, *personAfter) {
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
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				person := makePerson()
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
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
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

func TestPatch(t *testing.T) {
	t.Parallel()

	makePatchRequest := func(id any, body any) *http.Request {
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
			fmt.Sprintf("/person/%s", id),
			reader,
		)
		request.Header.Set("Content-Type", "application/json")
		return request
	}

	testCases := []testCase{
		{
			name: "not found",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				newAge := 60
				return makePatchRequest(uuid.New(), api.PersonPatchJSONRequestBody{
					Age: &newAge,
				}), nil
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid ID",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				newAge := 60
				return makePatchRequest("quux", api.PersonPatchJSONRequestBody{
						Age: &newAge,
					}), func(response *http.Response) {
						checkStringBody(t, response, regexp.MustCompile(`Invalid format for parameter personID`))
					}
			},
			status: http.StatusBadRequest,
		},
		{
			name: "valid age update",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				person := makePerson()
				personID, err := people.Create(context.Background(), &person)
				if err != nil {
					t.Fatalf("error initializing repo: %v", err)
				}
				newAge := person.Age / 2
				request := makePatchRequest(personID, api.PersonPatchJSONRequestBody{
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
	if personFilter.NameFragment != nil {
		values.Add("name", *personFilter.NameFragment)
	}
	if personFilter.PatronymicFragment != nil {
		values.Add("patronymic", *personFilter.PatronymicFragment)
	}
	if personFilter.SurnameFragment != nil {
		values.Add("surname", *personFilter.SurnameFragment)
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

func TestList(t *testing.T) {
	t.Parallel()

	fillDB := func(people repo.PersonRepo) {
		for i := 0; i < 100; i++ {
			person := makePerson()
			if _, err := people.Create(context.Background(), &person); err != nil {
				t.Fatalf("error initializing repo: %v", err)
			}
		}
	}

	testCases := []testCase{
		{
			name: "valid",
			init: func(t *testing.T, people repo.PersonRepo) (*http.Request, func(response *http.Response)) {
				minAge := 20
				nameFragment := "a"
				personFilter := domain.PersonFilter{AgeMin: &minAge, NameFragment: &nameFragment}
				paginationFilter := domain.PaginationFilter{Limit: 2, Offset: 1}

				var expected domain.Page[*domain.Person]
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
