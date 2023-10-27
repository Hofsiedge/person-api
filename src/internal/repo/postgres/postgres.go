package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hofsiedge/person-api/internal/config"
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/repo"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sentient errors
var (
	ErrPostgresError            = errors.New("postgres error")
	ErrCheckContsraintViolation = fmt.Errorf(
		"%w: check constraint violation", ErrPostgresError)
)

// pgxpool.Pool / github.com/jackc/pgx/v5
type PgxPoolInterface interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
}

type People struct {
	db PgxPoolInterface
}

// this function should not be used outside tests
func PeopleFromPgxPoolInterface(db PgxPoolInterface) *People {
	return &People{db}
}

func New(cfg config.DBConfig) (*People, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repo.ErrArgument, err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repo.ErrConnect, err)
	}

	return &People{pool}, nil
}

func (p *People) Close() {
	p.db.Close()
}

// ensure that People implements repo.PersonRepo
var _ repo.PersonRepo = &People{nil}

// Create implements repo.PersonRepo.
func (p *People) Create(ctx context.Context, person domain.Person) (uuid.UUID, error) {
	var personID pgtype.UUID

	row := p.db.QueryRow(ctx, `
		select people.create_person(
			name_ => $1, surname_ => $2, patronymic_ => $3,
			age_ => $4, sex_ => $5, nationality_ => $6)
		`,
		person.Name, person.Surname, person.Patronymic,
		person.Age, person.Sex, person.Nationality,
	)

	if err := row.Scan(&personID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation {
			return uuid.UUID{}, fmt.Errorf(
				"%w: failed a check constraint: %w", repo.ErrArgument, pgErr)
		}

		return uuid.UUID{}, fmt.Errorf("%w: %w", repo.ErrUnexpected, err)
	}

	if !personID.Valid {
		return uuid.UUID{}, fmt.Errorf(
			"%w: todo.create_task returned NULL", repo.ErrUnexpected)
	}

	return personID.Bytes, nil
}

// Delete implements repo.PersonRepo.
func (*People) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FullUpdate implements repo.PersonRepo.
func (*People) FullUpdate(ctx context.Context, id uuid.UUID, replacement domain.Person) error {
	panic("unimplemented")
}

// GetByID implements repo.PersonRepo.
func (*People) GetByID(ctx context.Context, id uuid.UUID) (domain.Person, error) {
	panic("unimplemented")
}

// List implements repo.PersonRepo.
func (*People) List(
	ctx context.Context, filter domain.PersonFilter, pagination domain.PaginationFilter,
) (domain.Page[domain.Person], error) {
	panic("unimplemented")
}

// PartialUpdate implements repo.PersonRepo.
func (*People) PartialUpdate(ctx context.Context, id uuid.UUID, partial domain.PersonPartial) error {
	panic("unimplemented")
}
