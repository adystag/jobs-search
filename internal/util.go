package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrHashMismatched = errors.New("hash mismatched")

type Option[T OptionConstraint] func(opt *T)

type OptionConstraint interface {
	JobsListerOption
}

type Hasher interface {
	Hash(plain string) (string, error)
}

type Comparator interface {
	Compare(hashed, plain string) error
}

type Timer interface {
	Now() time.Time
}

type Validator[T any] interface {
	Validate(ctx context.Context, val T) error
}

func ApplyOptions[T OptionConstraint](opt *T, opts ...Option[T]) {
	for _, fn := range opts {
		fn(opt)
	}
}

type ValidationError struct {
	field string
	tag   string
}

func (ve ValidationError) Field() string {
	return ve.field
}

func (ve ValidationError) Tag() string {
	return ve.tag
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s field validation failed at %s tag", ve.field, ve.tag)
}

func NewValidationError(field, tag string) ValidationError {
	return ValidationError{
		field: field,
		tag:   tag,
	}
}

type bcryptHasher struct {
	cost int
}

func (h bcryptHasher) Hash(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword(bytes.NewBufferString(plain).Bytes(), h.cost)
	if err != nil {
		return "", fmt.Errorf("generating bcrypt hash from password: %w", err)
	}

	return bytes.NewBuffer(hashed).String(), nil
}

func (h bcryptHasher) Compare(hashed, plain string) error {
	err := bcrypt.CompareHashAndPassword(
		bytes.NewBufferString(hashed).Bytes(),
		bytes.NewBufferString(plain).Bytes(),
	)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			err = ErrHashMismatched
		}

		return fmt.Errorf("comparing bcrypt hash with plain: %w", err)
	}

	return nil
}

func NewBcryptHasher(cost int) *bcryptHasher {
	return &bcryptHasher{
		cost: cost,
	}
}

type timer struct{}

func (timer) Now() time.Time {
	return time.Now()
}

func NewTimer() *timer {
	return &timer{}
}

type validationAggregator[T any] struct {
	validators []Validator[T]
}

func (v validationAggregator[T]) Validate(ctx context.Context, val T) error {
	for index, validator := range v.validators {
		err := validator.Validate(ctx, val)
		if err != nil {
			return fmt.Errorf("calling validator number %d: %w", index, err)
		}
	}

	return nil
}

func NewValidationAggregator[T any](v1, v2 Validator[T], vn ...Validator[T]) *validationAggregator[T] {
	return &validationAggregator[T]{
		validators: append([]Validator[T]{v1, v2}, vn...),
	}
}
