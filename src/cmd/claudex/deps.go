package main

import (
	"claudex/internal/services/clock"
	"claudex/internal/services/commander"
	"claudex/internal/services/env"
	"claudex/internal/services/uuid"

	"github.com/spf13/afero"
)

// Package-level default instances for production use
var (
	AppCmd   = commander.New()
	AppClock = clock.New()
	AppUUID  = uuid.New()
	AppEnv   = env.New()
	AppFs    = afero.NewOsFs()
)
