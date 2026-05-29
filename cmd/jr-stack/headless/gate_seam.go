package headless

import (
	"context"

	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// detectDepsForFn is the function used by RunHeadless to detect a given set of
// dependencies. It defaults to system.DetectDepsFor and can be replaced in
// tests via SetDetectDepsForFn.
var detectDepsForFn = func(ctx context.Context, deps []system.Dependency) system.DependencyReport {
	return system.DetectDepsFor(ctx, deps)
}

// SetDetectDepsForFn replaces the dependency-detection function used by
// RunHeadless for testing. It returns a restore function that resets the
// original value.
func SetDetectDepsForFn(fn func(ctx context.Context, deps []system.Dependency) system.DependencyReport) (restore func()) {
	old := detectDepsForFn
	detectDepsForFn = fn
	return func() { detectDepsForFn = old }
}
