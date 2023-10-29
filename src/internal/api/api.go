package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Hofsiedge/person-api/internal/completer"
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/filler"
	"github.com/Hofsiedge/person-api/internal/repo"
)

//go:generate oapi-codegen --config=types.cfg.yaml  ../../../openapi.yaml
//go:generate oapi-codegen --config=server.cfg.yaml ../../../openapi.yaml
//go:generate oapi-codegen --config=spec.cfg.yaml   ../../../openapi.yaml

// ensure that Server implements StrictServerInterface
var _ StrictServerInterface = &Server{
	People:    nil,
	Completer: nil,
	Logger:    nil,
}

var ErrInit = errors.New("unexpected nil in argument list")

// implements StrictServerInterface.
type Server struct {
	People    repo.PersonRepo
	Completer Completer
	Logger    *slog.Logger
}

type Completer interface {
	Complete(name string) (completer.CompletionData, error)
	UnlockingTime() (time.Time, error)
}

func New(repo repo.PersonRepo, completer Completer, logger *slog.Logger) (*Server, error) {
	if repo == nil || logger == nil {
		return nil, ErrInit
	}

	return &Server{repo, completer, logger}, nil
}

// PersonGet implements StrictServerInterface.
func (s *Server) PersonGet( //nolint:ireturn
	ctx context.Context, request PersonGetRequestObject,
) (PersonGetResponseObject, error) {
	person, err := s.People.GetByID(ctx, request.PersonID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return PersonGet404Response{}, nil
		case errors.Is(err, repo.ErrUnexpected):
			fallthrough
		default:
			s.Logger.Log(ctx, slog.LevelError, "unexpected error",
				slog.String("error", err.Error()))

			return PersonGet5XXResponse{http.StatusInternalServerError}, nil
		}
	}

	return PersonGet200JSONResponse{
		Age:         person.Age,
		Id:          person.ID,
		Name:        person.Name,
		Nationality: string(person.Nationality),
		Patronymic:  person.Patronymic,
		Sex:         Sex(person.Sex),
		Surname:     person.Surname,
	}, nil
}

// PersonList implements StrictServerInterface.
func (s *Server) PersonList( //nolint:ireturn
	ctx context.Context, request PersonListRequestObject,
) (PersonListResponseObject, error) {
	// TODO: test with nulls for paginaion
	// TODO: test with incorrect Nationality and Sex parameters
	page, err := s.People.List(ctx, domain.PersonFilter{
		Name:        request.Params.Name,
		Surname:     request.Params.Surname,
		Patronymic:  request.Params.Patronymic,
		Nationality: (*domain.Nationality)(request.Params.Nationality),
		Sex:         (*domain.Sex)(request.Params.Sex),
		AgeMin:      request.Params.AgeMin,
		AgeMax:      request.Params.AgeMax,
		Threshold:   request.Params.Threshold,
	}, domain.PaginationFilter{
		Offset: *request.Params.Offset,
		Limit:  *request.Params.Limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrArgument):
			return PersonList400Response{}, nil
		case errors.Is(err, repo.ErrUnexpected):
			fallthrough
		default:
			return PersonList5XXResponse{http.StatusInternalServerError}, nil
		}
	}

	people := make([]PersonFullWithID, len(page.Items))
	for i, person := range page.Items {
		people[i] = PersonFullWithID{
			Age:         person.Age,
			Id:          person.ID,
			Name:        person.Name,
			Nationality: string(person.Nationality),
			Patronymic:  person.Patronymic,
			Sex:         Sex(person.Sex),
			Surname:     person.Surname,
		}
	}

	return PersonList200JSONResponse{
		Pagination: PaginationOffsetLimit{
			CurrentLimit:  page.CurrentLimit,
			CurrentOffset: page.CurrentOffset,
			TotalItems:    page.TotalItems,
		},
		People: people,
	}, nil
}

// PersonPatch implements StrictServerInterface.
func (s *Server) PersonPatch( //nolint:ireturn
	ctx context.Context, request PersonPatchRequestObject,
) (PersonPatchResponseObject, error) {
	err := s.People.PartialUpdate(ctx, request.PersonID, domain.PersonPartial{
		Name:        request.Body.Name,
		Surname:     request.Body.Surname,
		Patronymic:  request.Body.Patronymic,
		Nationality: request.Body.Nationality,
		Sex:         (*domain.Sex)(request.Body.Sex),
		Age:         request.Body.Age,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrArgument):
			return PersonPatch400Response{}, nil
		case errors.Is(err, repo.ErrNotFound):
			return PersonPatch404Response{}, nil
		case errors.Is(err, repo.ErrUnexpected):
			fallthrough
		default:
			s.Logger.Log(ctx, slog.LevelError, "unexpected error",
				slog.String("error", err.Error()))

			return PersonPatch5XXResponse{http.StatusInternalServerError}, nil
		}
	}

	return PersonPatch200Response{}, nil
}

// PersonPost implements StrictServerInterface.
func (s *Server) PersonPost( //nolint:ireturn,cyclop
	ctx context.Context, request PersonPostRequestObject,
) (PersonPostResponseObject, error) {
	compData, err := s.Completer.Complete(request.Body.Name)
	if err != nil {
		switch {
		case errors.Is(err, filler.ErrUser):
			return PersonPost422Response{}, nil

		case errors.Is(err, filler.ErrEnvironment), errors.Is(err, filler.ErrAPI):
			return PersonPost5XXResponse{http.StatusInternalServerError}, nil

		case errors.Is(err, filler.ErrLimitReached):
			unlockingTime, err := s.Completer.UnlockingTime()
			if err != nil {
				return PersonPost5XXResponse{http.StatusInternalServerError}, nil //nolint:nilerr
			}

			return PersonPost503Response{
				Headers: PersonPost503ResponseHeaders{
					RetryAfter: int(time.Until(unlockingTime).Seconds()),
				},
			}, nil

		default:
			return PersonPost5XXResponse{http.StatusInternalServerError}, nil
		}
	}

	person := domain.Person{
		Name:        request.Body.Name,
		Surname:     request.Body.Surname,
		Patronymic:  request.Body.Patronymic,
		Nationality: compData.Nationality,
		Sex:         compData.Sex,
		Age:         compData.Age,
		ID:          [16]byte{},
	}

	personID, err := s.People.Create(ctx, person)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrArgument):
			return PersonPost400Response{}, nil
		case errors.Is(err, repo.ErrUnexpected):
			fallthrough
		default:
			s.Logger.Log(ctx, slog.LevelError, "unexpected error",
				slog.String("error", err.Error()))

			return PersonPost5XXResponse{http.StatusInternalServerError}, nil
		}
	}

	response := PersonPost201JSONResponse(personID)

	return response, nil
}

// PersonPut implements StrictServerInterface.
func (s *Server) PersonPut( //nolint:ireturn
	ctx context.Context, request PersonPutRequestObject,
) (PersonPutResponseObject, error) {
	err := s.People.FullUpdate(ctx, request.PersonID, domain.Person{
		Name:        request.Body.Name,
		Surname:     request.Body.Surname,
		Patronymic:  request.Body.Patronymic,
		Nationality: domain.Nationality(request.Body.Nationality),
		Sex:         domain.Sex(request.Body.Sex),
		Age:         request.Body.Age,
		ID:          [16]byte{},
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrArgument):
			return PersonPut400Response{}, nil
		case errors.Is(err, repo.ErrNotFound):
			return PersonPut404Response{}, nil
		case errors.Is(err, repo.ErrUnexpected):
			fallthrough
		default:
			s.Logger.Log(ctx, slog.LevelError, "unexpected error",
				slog.String("error", err.Error()))

			return PersonPut5XXResponse{http.StatusInternalServerError}, nil
		}
	}

	return PersonPut200Response{}, nil
}

// PersonDelete implements StrictServerInterface.
func (s *Server) PersonDelete( //nolint:ireturn
	ctx context.Context, request PersonDeleteRequestObject,
) (PersonDeleteResponseObject, error) {
	err := s.People.Delete(ctx, request.PersonID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrArgument):
			return PersonDelete400Response{}, nil
		case errors.Is(err, repo.ErrNotFound):
			return PersonDelete404Response{}, nil
		case errors.Is(err, repo.ErrUnexpected):
			fallthrough
		default:
			s.Logger.Log(ctx, slog.LevelError, "unexpected error",
				slog.String("error", err.Error()))

			return PersonDelete5XXResponse{http.StatusInternalServerError}, nil
		}
	}

	return PersonDelete200Response{}, nil
}
