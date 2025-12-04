package env

import "os"

// Environment abstracts environment variable access
type Environment interface {
	Get(key string) string
	Set(key, value string)
}

// OsEnv is the production implementation of Environment
type OsEnv struct{}

func (e *OsEnv) Get(key string) string {
	return os.Getenv(key)
}

func (e *OsEnv) Set(key, value string) {
	os.Setenv(key, value)
}

// New creates a new Environment instance
func New() Environment {
	return &OsEnv{}
}
