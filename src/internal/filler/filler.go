package filler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const GetRequestTimeout = time.Second * 3

var (
	ErrFiller = errors.New("filler error")

	ErrNotReady     = fmt.Errorf("%w: filler is not ready (not used yet)", ErrFiller)
	ErrInvalidURL   = fmt.Errorf("%w: invalid URL", ErrFiller)
	ErrTimeout      = fmt.Errorf("%w: timeout", ErrFiller)
	ErrNetworkError = fmt.Errorf("%w: network error", ErrFiller)

	ErrInvalidResponse = fmt.Errorf("%w: invalid response", ErrFiller)
	ErrInvalidHeader   = fmt.Errorf("%w: invalid header", ErrInvalidResponse)
	ErrInvalidStatus   = fmt.Errorf("%w: invalid status code", ErrInvalidResponse)

	ErrInvalidAPIToken = fmt.Errorf("%w: invalid API token", ErrFiller)
	ErrInvalidName     = fmt.Errorf("%w: invalid name", ErrFiller)
	ErrLimitReached    = fmt.Errorf("%w: request limit reached", ErrFiller)

	ErrConversion = fmt.Errorf("%w: conversion error", ErrFiller)

	ErrNotFound = fmt.Errorf("%w: not found", ErrFiller)
)

type Converter[T any] interface {
	Convert() (T, error)
}

type Filler[T any, C Converter[T]] struct {
	Client       *http.Client
	resetTime    time.Time
	token        *string
	baseURL      string
	requestsLeft int
	requestLimit int
	sync.RWMutex
	valid bool
}

// New returns a new Filler that accesses the API at baseURL with an optional API token
// and an optional http.Client.
//
// If token is nil, no token is used. If client is nil, a client with configured
// timeout is used.
func New[T any, C Converter[T]](baseURL string, token *string, client *http.Client) Filler[T, C] {
	if client == nil {
		//nolint:exhaustruct
		client = &http.Client{
			Timeout: GetRequestTimeout,
		}
	}
	//nolint:exhaustruct
	return Filler[T, C]{
		Client:  client,
		valid:   false,
		token:   token,
		baseURL: baseURL,
	}
}

// RequestsLeft returns the number of requests left until the rate limiter reset
func (f *Filler[_, _]) RequestsLeft() (int, error) {
	f.RLock()
	defer f.RUnlock()

	if f.valid {
		return f.requestsLeft, nil
	}

	return 0, ErrNotReady
}

// RequestLimit returns the number of requests permitted by the rate limiter
// for the time period
func (f *Filler[_, _]) RequestLimit() (int, error) {
	f.RLock()
	defer f.RUnlock()

	if f.valid {
		return f.requestLimit, nil
	}

	return 0, ErrNotReady
}

// ResetTime returns the time when rate limiter resets
func (f *Filler[_, _]) ResetTime() (time.Time, error) {
	f.RLock()
	defer f.RUnlock()

	if f.valid {
		return f.resetTime, nil
	}

	return time.Time{}, ErrNotReady
}

func parseHeader(response *http.Response, header string) (int, error) {
	textValue := response.Header.Get(header)

	value, err := strconv.Atoi(textValue)
	if err != nil {
		err = fmt.Errorf("%w: header %q, value %q", ErrInvalidHeader, header, textValue)

		return 0, err
	}

	return value, nil
}

// performRequest performs a GET request for the provided name
// and updated Filler fields using response headers
func (f *Filler[_, _]) performRequest(name string) (*http.Response, error) {
	values := url.Values{}

	values.Add("name", name)

	if f.token != nil {
		values.Add("apikey", *f.token)
	}

	URL := fmt.Sprintf("%s?%s", f.baseURL, values.Encode())

	response, err := f.Client.Get(URL) //nolint:noctx
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			return nil, ErrTimeout
		}

		return nil, fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	// read headers
	limit, err := parseHeader(response, "X-Rate-Limit-Limit")
	if err != nil {
		return nil, err
	}

	remaining, err := parseHeader(response, "X-Rate-Limit-Remaining")
	if err != nil {
		return nil, err
	}

	reset, err := parseHeader(response, "X-Rate-Limit-Reset")
	if err != nil {
		return nil, err
	}

	// update fields
	f.Lock()
	if f.valid {
		if remaining < f.requestsLeft {
			f.requestsLeft = remaining
		}
	} else {
		f.requestLimit = limit
		f.resetTime = time.Now().Add(time.Duration(reset) * time.Second)
		f.requestsLeft = remaining
		f.valid = true
	}
	f.Unlock()

	return response, err
}

func (f *Filler[T, C]) Fill(name string) (T, error) { //nolint:cyclop,ireturn
	var (
		validResponse C
		result        T
	)

	f.RLock()
	if f.valid && f.requestsLeft == 0 && f.resetTime.After(time.Now()) {
		return result, ErrLimitReached
	}
	f.RUnlock()

	response, err := f.performRequest(name)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusPaymentRequired:
		return result, ErrInvalidAPIToken

	case http.StatusUnprocessableEntity:
		return result, ErrInvalidName

	case http.StatusTooManyRequests:
		return result, ErrLimitReached

	case http.StatusOK:
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return result, fmt.Errorf("%w: %w", ErrNetworkError, err)
		}

		if err = json.Unmarshal(bytes, &validResponse); err != nil {
			return result, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
		}

		result, err = validResponse.Convert()
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return result, err //nolint:wrapcheck
			}

			return result, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
		}

		return result, nil

	default:
		return result, ErrInvalidStatus
	}
}
