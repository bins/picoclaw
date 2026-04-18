package gateway

import (
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/health"
)

type HealthExtension func(server *health.Server, cfg *config.Config, workspace string)

var healthExtensions []HealthExtension

func RegisterHealthExtension(fn HealthExtension) {
	healthExtensions = append(healthExtensions, fn)
}

func runHealthExtensions(server *health.Server, cfg *config.Config, workspace string) {
	for _, fn := range healthExtensions {
		fn(server, cfg, workspace)
	}
}
