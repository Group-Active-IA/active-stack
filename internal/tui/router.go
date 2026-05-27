package tui

// route holds the forward and backward neighbors for a screen.
type route struct {
	forward  Screen
	backward Screen
}

// linearRoutes defines the install flow navigation graph.
// Custom picker is conditional — the router skips it for Lite/Full
// (handled by Model.nextScreen).
var linearRoutes = map[Screen]route{
	ScreenWelcome:      {forward: ScreenDetection},
	ScreenDetection:    {forward: ScreenAgents, backward: ScreenWelcome},
	ScreenAgents:       {forward: ScreenMode, backward: ScreenDetection},
	ScreenMode:         {forward: ScreenCustomPicker, backward: ScreenAgents},
	ScreenCustomPicker: {forward: ScreenReview, backward: ScreenMode},
	ScreenReview:       {forward: ScreenInstalling, backward: ScreenMode},
	ScreenInstalling:   {forward: ScreenComplete, backward: ScreenReview},
	ScreenComplete:     {backward: ScreenInstalling},
}

// nextScreen returns the screen that follows current in the linear flow.
// ok is false if there is no forward route.
func nextScreen(current Screen) (Screen, bool) {
	r, ok := linearRoutes[current]
	if !ok || r.forward == ScreenUnknown {
		return ScreenUnknown, false
	}
	return r.forward, true
}

// prevScreen returns the screen that precedes current in the linear flow.
// ok is false if there is no backward route.
func prevScreen(current Screen) (Screen, bool) {
	r, ok := linearRoutes[current]
	if !ok || r.backward == ScreenUnknown {
		return ScreenUnknown, false
	}
	return r.backward, true
}
