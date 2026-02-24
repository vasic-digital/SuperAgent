//go:build !nohelixspecifier

package specifier

import (
	helixadapters "digital.vasic.helixspecifier/pkg/adapters"
	helixceremony "digital.vasic.helixspecifier/pkg/ceremony"
	helixcfg "digital.vasic.helixspecifier/pkg/config"
	helixengine "digital.vasic.helixspecifier/pkg/engine"
	helixgsd "digital.vasic.helixspecifier/pkg/gsd"
	helixintent "digital.vasic.helixspecifier/pkg/intent"
	helixmemory "digital.vasic.helixspecifier/pkg/memory"
	helixmetrics "digital.vasic.helixspecifier/pkg/metrics"
	helixspeckit "digital.vasic.helixspecifier/pkg/speckit"
	helixpowers "digital.vasic.helixspecifier/pkg/superpowers"
	"github.com/sirupsen/logrus"
)

// NewOptimalSpecAdapter creates a fully-configured HelixSpecifier adapter.
// This is the DEFAULT factory (active without any build tags).
// To opt out, build with: go build -tags nohelixspecifier
func NewOptimalSpecAdapter() *SpecAdapter {
	cfg := helixcfg.FromEnv()
	logger := logrus.StandardLogger()

	// Create core engine
	engine := helixengine.New(cfg, logger)

	// Create and register SpecKit pillar
	sk := helixspeckit.NewPillar(cfg, logger)
	engine.RegisterSpecKit(sk)

	// Create and register Superpowers pillar
	sp := helixpowers.NewPillar(cfg, logger)
	engine.RegisterSuperpowers(sp)

	// Create and register GSD pillar
	gsd := helixgsd.NewPillar(logger)
	engine.RegisterGSD(gsd)

	// Create and register ceremony scaler
	cs := helixceremony.NewScaler(cfg, logger)
	engine.RegisterCeremonyScaler(cs)

	// Create and register spec memory
	mem := helixmemory.NewStore()
	engine.RegisterSpecMemory(mem)

	// Register well-known CLI agent adapters
	for name, adapter := range helixadapters.WellKnownAdapters() {
		engine.RegisterAdapter(name, adapter)
	}

	// Create metrics (for monitoring)
	_ = helixmetrics.NewMetrics()

	// Create intent classifier and wire into engine
	ic := helixintent.NewClassifier()
	engine.RegisterClassifier(ic.Classify)

	return NewSpecAdapter(engine)
}

// IsHelixSpecifierEnabled returns true -- HelixSpecifier is the default.
func IsHelixSpecifierEnabled() bool {
	return true
}

// SpecifierBackendName returns the module name.
func SpecifierBackendName() string {
	return "digital.vasic.helixspecifier"
}
