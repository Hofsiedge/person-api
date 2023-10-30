package completer

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Hofsiedge/person-api/internal/config"
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/filler/agify"
	"github.com/Hofsiedge/person-api/internal/filler/genderize"
	"github.com/Hofsiedge/person-api/internal/filler/nationalize"
)

const Timeout = time.Second * 10

var (
	ErrCompleterError = errors.New("completer error")

	ErrUserError     = fmt.Errorf("%w: user error", ErrCompleterError)
	ErrInvalidConfig = fmt.Errorf("%w: config error", ErrCompleterError)

	ErrWrongUsage = fmt.Errorf("%w: wrong usage", ErrCompleterError)
)

// Completer is an abstraction over filler.Filler that uses multiple concurrently
// fillers and combines their results
type Completer struct {
	client *http.Client
	// time when the fillers request quota resets
	unlockTime   *time.Time
	genderizer   genderize.Genderizer
	nationalizer nationalize.Nationalizer
	agifier      agify.Agifier
}

type CompletionData struct {
	Sex         domain.Sex
	Nationality domain.Nationality
	Age         int
}

func New(cfg config.CompleterConfig, client *http.Client) *Completer {
	if client == nil {
		//nolint:exhaustruct
		client = &http.Client{
			Timeout: Timeout,
		}
	}

	var token *string
	if len(cfg.CompleterToken) > 0 {
		token = &cfg.CompleterToken
	}

	return &Completer{
		client:       client,
		unlockTime:   nil,
		genderizer:   genderize.New(cfg.GenderizeURL, token, client),
		nationalizer: nationalize.New(cfg.NationalizeURL, token, client),
		agifier:      agify.New(cfg.AgifyURL, token, client),
	}
}

func bToI(b bool) int {
	if b {
		return 1
	}

	return 0
}

func (c *Completer) Complete(name string) (CompletionData, error) {
	var wg sync.WaitGroup //nolint:varnamelen

	wg.Add(3) //nolint:gomnd

	var (
		data                           CompletionData
		sexErr, nationalityErr, ageErr error
	)

	go func() {
		data.Sex, sexErr = c.genderizer.Fill(name)

		wg.Done()
	}()

	go func() {
		data.Nationality, nationalityErr = c.nationalizer.Fill(name)

		wg.Done()
	}()

	go func() {
		data.Age, ageErr = c.agifier.Fill(name)

		wg.Done()
	}()

	wg.Wait()

	errorCount := bToI(sexErr != nil) +
		bToI(nationalityErr != nil) +
		bToI(ageErr != nil)

	if errorCount == 0 {
		return data, nil
	}

	//nolint:stylecheck
	err := fmt.Errorf("%w (%d fillers failed). filler errors:", ErrCompleterError, errorCount)
	if ageErr != nil {
		err = fmt.Errorf("%w {age: %w}", err, ageErr)
	}

	if nationalityErr != nil {
		err = fmt.Errorf("%w {nationality: %w}", err, nationalityErr)
	}

	if sexErr != nil {
		err = fmt.Errorf("%w {sex: %w}", err, sexErr)
	}

	return data, err
}

//nolint:varnamelen
func (c *Completer) updateUnlockTime() error {
	t1, err1 := c.agifier.ResetTime()
	t2, err2 := c.genderizer.ResetTime()
	t3, err3 := c.nationalizer.ResetTime()

	if err1 != nil || err2 != nil || err3 != nil {
		return fmt.Errorf("%w: some of the fillers are not ready", ErrWrongUsage)
	}

	if t1.Before(t2) {
		t1 = t2
	}

	if t1.Before(t3) {
		t1 = t3
	}

	c.unlockTime = &t1

	return nil
}

func (c *Completer) UnlockingTime() (time.Time, error) {
	if c.unlockTime == nil {
		if err := c.updateUnlockTime(); err != nil {
			return time.Time{}, err
		}
	}

	return *c.unlockTime, nil
}
