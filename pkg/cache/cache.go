package cache

import (
	"time"
)

type ICache interface {
	WriteCache(string, string) error
	ReadCache(string) (string, error)
	Expire(string, time.Duration) error
	RemoveCache(string) error
	KeyExists(string) bool
	WriteCacheWithTTL(string, string, time.Duration) error
	Name() string
}
