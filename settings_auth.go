package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"slices"
	"sync"
	"time"
)

type SettingsAuth struct {
	password string
	salts    *ExpiringSet[string]
}

func NewSettingsAuth(password string) *SettingsAuth {
	return &SettingsAuth{
		password: password,
		salts:    NewExpiringSet[string](10, 10*time.Second),
	}
}

func (auth *SettingsAuth) GenerateSalt() string {
	saltData := make([]byte, 16)
	rand.Read(saltData)
	salt := fmt.Sprintf("%x", saltData)

	auth.salts.Add(salt)
	return salt
}

func (auth *SettingsAuth) CheckPassword(hash string, salt string) bool {
	if auth.password == "" {
		return false // disable settings updater
	}
	if !auth.salts.Remove(salt) {
		return false
	}
	serverHashData := sha256.Sum256([]byte(salt + auth.password + salt))
	serverHash := fmt.Sprintf("%x", serverHashData)
	return hash == serverHash
}

type expiringSetItem[T comparable] struct {
	value     T
	expiresAt time.Time
}

type ExpiringSet[T comparable] struct {
	mu      sync.Mutex
	items   []expiringSetItem[T]
	maxSize int
	ttl     time.Duration
}

func NewExpiringSet[T comparable](maxSize int, ttl time.Duration) *ExpiringSet[T] {
	return &ExpiringSet[T]{
		items:   make([]expiringSetItem[T], 0),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

func (s *ExpiringSet[T]) Add(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.items) >= s.maxSize {
		s.items = s.items[1:]
	}

	s.items = slices.DeleteFunc(s.items, func(item expiringSetItem[T]) bool {
		return item.value == value
	})

	s.items = append(s.items, expiringSetItem[T]{
		value:     value,
		expiresAt: time.Now().Add(s.ttl),
	})
}

func (s *ExpiringSet[T]) Remove(value T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	i := slices.IndexFunc(s.items, func(item expiringSetItem[T]) bool {
		return item.value == value && item.expiresAt.After(now)
	})

	if i == -1 {
		return false
	}

	s.items = slices.Delete(s.items, i, i+1)
	return true
}
