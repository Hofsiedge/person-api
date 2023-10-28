package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/google/uuid"
)

// sentinel errors.
var (
	ErrRepo       = errors.New("repo error")
	ErrConnect    = fmt.Errorf("%w: could not connect", ErrRepo)
	ErrDisconnect = fmt.Errorf("%w: error closing connection", ErrRepo)
	ErrNotFound   = fmt.Errorf("%w: not found", ErrRepo)
	ErrUnexpected = fmt.Errorf("%w: unexpected error", ErrRepo)
	ErrArgument   = fmt.Errorf("%w: argument error", ErrRepo)
)

type WithID[I comparable] interface {
	GetID() I
}

// T - main type, I - ID type, P - partial type, F - filter type.
type Repo[T WithID[I], I comparable, P any, F any] interface {
	Create(ctx context.Context, obj T) (I, error)
	List(ctx context.Context, filter F, pagination domain.PaginationFilter) (domain.Page[T], error)
	GetByID(ctx context.Context, id I) (T, error)
	PartialUpdate(ctx context.Context, id I, partial P) error
	FullUpdate(ctx context.Context, id I, replacement T) error
	Delete(ctx context.Context, id I) error
}

type PersonRepo Repo[domain.Person, uuid.UUID, domain.PersonPartial, domain.PersonFilter]
