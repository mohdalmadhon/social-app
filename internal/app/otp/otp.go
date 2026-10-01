package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

const (
	CodeLength      = 6
	MaxAttempts     = 3
	CodeTTL         = 10 * time.Minute
	Cooldown        = 60 * time.Second
	TokenTTL        = 15 * time.Minute
	MaxSendsPerHour = 5
)

var (
	ErrNoCode    = errors.New("no code")
	ErrExpired   = errors.New("code expired")
	ErrLocked    = errors.New("too many attempts")
	ErrWrongCode = errors.New("wrong code")
)

type RateLimitError struct {
	RetryAfter int
	Message    string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

type entry struct {
	hash         [32]byte
	expires      time.Time
	attemptsLeft int
	lastSent     time.Time
	sends        []time.Time
}

type token struct {
	email   string
	expires time.Time
}

type Store struct {
	mu     sync.Mutex
	codes  map[string]*entry
	tokens map[string]token
}

func NewStore() *Store {
	return &Store{
		codes:  make(map[string]*entry),
		tokens: make(map[string]token),
	}
}

func hashCode(email, code string) [32]byte {
	return sha256.Sum256([]byte(email + ":" + code))
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (s *Store) purge(now time.Time) {
	for email, e := range s.codes {
		if now.After(e.expires) && now.Sub(e.lastSent) > time.Hour {
			delete(s.codes, email)
		}
	}

	for value, t := range s.tokens {
		if now.After(t.expires) {
			delete(s.tokens, value)
		}
	}
}

func (s *Store) Issue(email string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.purge(now)

	e := s.codes[email]

	if e != nil {
		if wait := Cooldown - now.Sub(e.lastSent); wait > 0 {
			seconds := int(wait.Seconds()) + 1

			return "", &RateLimitError{
				RetryAfter: seconds,
				Message:    fmt.Sprintf("Please wait %ds before requesting another code", seconds),
			}
		}

		recent := e.sends[:0]

		for _, sentAt := range e.sends {
			if now.Sub(sentAt) < time.Hour {
				recent = append(recent, sentAt)
			}
		}

		e.sends = recent

		if len(e.sends) >= MaxSendsPerHour {
			wait := time.Hour - now.Sub(e.sends[0])
			seconds := int(wait.Seconds()) + 1

			return "", &RateLimitError{
				RetryAfter: seconds,
				Message:    "Too many codes requested. Try again later",
			}
		}
	} else {
		e = &entry{}
		s.codes[email] = e
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	code := fmt.Sprintf("%06d", n.Int64())

	e.hash = hashCode(email, code)
	e.expires = now.Add(CodeTTL)
	e.attemptsLeft = MaxAttempts
	e.lastSent = now
	e.sends = append(e.sends, now)

	return code, nil
}

func (s *Store) Forget(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.codes, email)
}

func (s *Store) Verify(email, code string) (string, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	e := s.codes[email]

	if e == nil {
		return "", 0, ErrNoCode
	}

	if now.After(e.expires) {
		e.attemptsLeft = 0
		return "", 0, ErrExpired
	}

	if e.attemptsLeft <= 0 {
		return "", 0, ErrLocked
	}

	given := hashCode(email, code)

	if subtle.ConstantTimeCompare(given[:], e.hash[:]) != 1 {
		e.attemptsLeft--

		if e.attemptsLeft <= 0 {
			return "", 0, ErrLocked
		}

		return "", e.attemptsLeft, ErrWrongCode
	}

	e.attemptsLeft = 0
	e.expires = now

	value, err := randomHex(32)
	if err != nil {
		return "", 0, err
	}

	s.tokens[value] = token{
		email:   email,
		expires: now.Add(TokenTTL),
	}

	return value, 0, nil
}

func (s *Store) CheckToken(value, email string) bool {
	if value == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tokens[value]
	if !ok {
		return false
	}

	if time.Now().After(t.expires) {
		delete(s.tokens, value)
		return false
	}

	return t.email == email
}

func (s *Store) DeleteToken(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, value)
}
