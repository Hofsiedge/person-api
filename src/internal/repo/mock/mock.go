package mock

import (
	"context"
	"reflect"
	"slices"
	"strings"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/repo"
	"github.com/google/uuid"
)

type People struct {
	People map[uuid.UUID]domain.Person
}

// ensure People implements the interface
var _ repo.PersonRepo = &People{
	People: nil,
}

func New() *People {
	return &People{
		make(map[uuid.UUID]domain.Person),
	}
}

// Create implements repo.Repo.
func (p *People) Create(ctx context.Context, obj *domain.Person) (uuid.UUID, error) {
	if obj == nil {
		return uuid.UUID{}, repo.ErrArgument
	}

	id := uuid.New()
	obj.ID = id
	p.People[id] = *obj

	return id, nil
}

// Delete implements repo.Repo.
func (p *People) Delete(ctx context.Context, id uuid.UUID) error {
	if _, found := p.People[id]; !found {
		return repo.ErrNotFound
	}

	delete(p.People, id)

	return nil
}

// FullUpdate implements repo.Repo.
func (p *People) FullUpdate(ctx context.Context, personID uuid.UUID, replacement *domain.Person) error {
	if _, found := p.People[personID]; !found {
		return repo.ErrNotFound
	}

	task := *replacement
	task.ID = personID
	p.People[personID] = task

	return nil
}

// GetByID implements repo.Repo.
func (p *People) GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	task, found := p.People[id]
	if !found {
		return nil, repo.ErrNotFound
	}

	return &task, nil
}

//nolint:cyclop
func personMatches(filter domain.PersonFilter, person domain.Person) bool {
	// filter condition violations
	youngerThanMinAge := (filter.AgeMin != nil) && (*filter.AgeMin > person.Age)
	olderThanMaxAge := (filter.AgeMax != nil) && (*filter.AgeMax < person.Age)
	nameMismatch := (filter.NameFragment != nil) &&
		!strings.Contains(person.Name, *filter.NameFragment)

	surnameMismatch := (filter.SurnameFragment != nil) &&
		!strings.Contains(person.Surname, *filter.SurnameFragment)

	patronymicMismatch := (filter.PatronymicFragment != nil) &&
		((*filter.PatronymicFragment == "") && (person.Patronymic != "") ||
			(*filter.PatronymicFragment != "") &&
				!strings.Contains(person.Patronymic, *filter.PatronymicFragment))

	nationalityMismatch := (filter.Nationality != nil) &&
		(*filter.Nationality != person.Nationality)

	sexMismatch := (filter.Sex != nil) && (*filter.Sex != person.Sex)

	return !(youngerThanMinAge || olderThanMaxAge || nameMismatch || surnameMismatch ||
		patronymicMismatch || nationalityMismatch || sexMismatch)
}

// List implements repo.Repo.
func (p *People) List(
	ctx context.Context, filter domain.PersonFilter, pagination domain.PaginationFilter,
) (domain.Page[*domain.Person], error) {
	keys := make([]uuid.UUID, 0)
	for key := range p.People {
		keys = append(keys, key)
	}

	slices.SortFunc(keys, func(a, b uuid.UUID) int {
		return strings.Compare(a.String(), b.String())
	})

	result := make([]*domain.Person, 0)
	recordIndex := 0

	for _, id := range keys {
		person := p.People[id]

		if !personMatches(filter, person) {
			continue
		}

		if (recordIndex >= pagination.Offset) && (recordIndex < pagination.Offset+pagination.Limit) {
			result = append(result, &person)
		}
		recordIndex++
	}

	page := domain.Page[*domain.Person]{
		Items:         result,
		CurrentLimit:  pagination.Limit,
		CurrentOffset: pagination.Offset,
		TotalItems:    len(p.People),
	}

	return page, nil
}

// PartialUpdate implements repo.Repo.
func (p *People) PartialUpdate(ctx context.Context, personID uuid.UUID, partial domain.PersonPartial) error {
	person, found := p.People[personID]
	if !found {
		return repo.ErrNotFound
	}

	if (partial == domain.PersonPartial{
		Name:        nil,
		Surname:     nil,
		Patronymic:  nil,
		Nationality: nil,
		Sex:         nil,
		Age:         nil,
	}) {
		return repo.ErrArgument
	}
	// storing (*T)(nil) in `any` is dangerous (!= nil)
	// but this is just mock code, so reflect.IsNil is used to deal with
	// the issue
	values := map[any]any{
		&person.Name:        partial.Name,
		&person.Surname:     partial.Surname,
		&person.Patronymic:  partial.Patronymic,
		&person.Nationality: partial.Nationality,
		&person.Sex:         partial.Sex,
		&person.Age:         partial.Age,
	}
	for old, value := range values {
		if !reflect.ValueOf(value).IsNil() {
			left := reflect.ValueOf(old).Elem()
			right := reflect.ValueOf(value).Elem()
			left.Set(right)
		}
	}

	p.People[personID] = person

	return nil
}
