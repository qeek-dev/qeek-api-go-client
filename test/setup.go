package test

import (
	"testing"
)

type Env struct {
	T *testing.T
	B *testing.B
}

func SetupEnv(t *testing.T) *Env {
	env := new(Env)
	return env
}

func BenchmarkSetupEnv(b *testing.B) *Env {
	env := new(Env)
	return env
}
