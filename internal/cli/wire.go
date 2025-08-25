//go:build wireinject
// +build wireinject

package cli

import (
	"github.com/google/wire"
	"github.com/nkaewam/taskw/internal/cli/clean"
	"github.com/nkaewam/taskw/internal/cli/file"
	"github.com/nkaewam/taskw/internal/cli/generation"
	"github.com/nkaewam/taskw/internal/cli/project"
	"github.com/nkaewam/taskw/internal/cli/scan"
	"github.com/nkaewam/taskw/internal/cli/ui"
	"github.com/nkaewam/taskw/internal/config"
)

// Container holds all the injected services
type Container struct {
	UI         ui.Service
	Project    project.Service
	Scan       scan.Service
	Generation generation.Service
	Clean      clean.Service
	File       file.Service
	Config     *config.Config
}

// ProviderSet is the Wire provider set for all CLI services
var ProviderSet = wire.NewSet(
	GeneratedProviderSet,
)

// InitializeContainer initializes the dependency injection container
func InitializeContainer(configPath string) (*Container, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Container), "*"),
	)
	return &Container{}, nil
}
