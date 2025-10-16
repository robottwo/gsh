package environment

import (
	"os"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

// DynamicEnviron implements expand.Environ to provide a dynamic environment
// that includes both system environment variables and GSH-specific variables
type DynamicEnviron struct {
	systemEnv expand.Environ
	gshVars   map[string]string
}

// NewDynamicEnviron creates a new DynamicEnviron that wraps the system environment
// NewDynamicEnviron creates a DynamicEnviron that wraps the current OS environment
// and an empty set of GSH-specific variables.
func NewDynamicEnviron() *DynamicEnviron {
	return &DynamicEnviron{
		systemEnv: expand.ListEnviron(os.Environ()...),
		gshVars:   make(map[string]string),
	}
}

// Get retrieves a variable by name, checking GSH variables first, then system environment
func (de *DynamicEnviron) Get(name string) expand.Variable {
	// Check GSH variables first
	if value, exists := de.gshVars[name]; exists {
		return expand.Variable{
			Exported: true,
			Kind:     expand.String,
			Str:      value,
		}
	}

	// Fall back to system environment
	return de.systemEnv.Get(name)
}

// Each iterates over all variables, including both GSH and system variables
func (de *DynamicEnviron) Each(fn func(name string, vr expand.Variable) bool) {
	// First, iterate over GSH variables
	for name, value := range de.gshVars {
		if !fn(name, expand.Variable{
			Exported: true,
			Kind:     expand.String,
			Str:      value,
		}) {
			return
		}
	}

	// Then iterate over system environment, skipping GSH variables we already added
	de.systemEnv.Each(func(name string, vr expand.Variable) bool {
		if _, isGSH := de.gshVars[name]; !isGSH {
			return fn(name, vr)
		}
		return true
	})
}

// UpdateGSHVar updates a GSH variable in the dynamic environment
func (de *DynamicEnviron) UpdateGSHVar(name, value string) {
	de.gshVars[name] = value
}

// UpdateSystemEnv updates the system environment wrapper
func (de *DynamicEnviron) UpdateSystemEnv() {
	de.systemEnv = expand.ListEnviron(os.Environ()...)
}

// SyncVariablesToEnv syncs gsh's internal variables to system environment variables
// SyncVariablesToEnv makes GSH-specific runner variables visible to the OS environment
// and ensures the runner uses a DynamicEnviron that overlays those variables.
//
// If the runner already has a DynamicEnviron it is reused; otherwise a new one is created.
// For a predefined set of GSH variable names, the function copies any values present in
// runner.Vars into the process environment and into the DynamicEnviron's GSH variable map.
// After updating GSH variables it refreshes the DynamicEnviron's view of the system
// environment and assigns the DynamicEnviron to runner.Env.
func SyncVariablesToEnv(runner *interp.Runner) {
	// Check if we already have a DynamicEnviron, if not create one
	var dynamicEnv *DynamicEnviron
	if existingDynamicEnv, ok := runner.Env.(*DynamicEnviron); ok {
		dynamicEnv = existingDynamicEnv
	} else {
		dynamicEnv = NewDynamicEnviron()
	}

	gshVars := []string{
		"GSH_PROMPT", "GSH_APROMPT", "GSH_LOG_LEVEL", "GSH_CLEAN_LOG_FILE",
		"GSH_MINIMUM_HEIGHT", "GSH_FAST_MODEL_API_KEY", "GSH_FAST_MODEL_BASE_URL",
		"GSH_FAST_MODEL_ID", "GSH_SLOW_MODEL_API_KEY", "GSH_SLOW_MODEL_BASE_URL",
		"GSH_SLOW_MODEL_ID", "GSH_CONTEXT_TYPES_FOR_AGENT", "GSH_AGENT_CONTEXT_WINDOW_TOKENS",
		"GSH_AGENT_APPROVED_BASH_COMMAND_REGEX", "GSH_AGENT_MACROS",
	}

	for _, varName := range gshVars {
		if varValue, exists := runner.Vars[varName]; exists {
			value := varValue.String()

			os.Setenv(varName, value)
			dynamicEnv.UpdateGSHVar(varName, value)
		}
	}

	// Update the system environment in the dynamic environment
	dynamicEnv.UpdateSystemEnv()

	// Set the runner's environment to our dynamic environment
	runner.Env = dynamicEnv
}

// SyncVariableToEnv synchronizes the named GSH variable from the runner's Vars into the process environment
// and updates the runner's DynamicEnviron entry for that variable when the runner's Env is a DynamicEnviron.
// If the variable is not present in runner.Vars, the function does nothing.
func SyncVariableToEnv(runner *interp.Runner, varName string) {
	if varValue, exists := runner.Vars[varName]; exists {
		value := varValue.String()
		os.Setenv(varName, value)

		// Update in the dynamic environment
		if dynamicEnv, ok := runner.Env.(*DynamicEnviron); ok {
			dynamicEnv.UpdateGSHVar(varName, value)
		}
	}
}

// IsGSHVariable reports whether the given name is a GSH-specific variable that should be synchronized to the environment.
// Known GSH variable names and any name beginning with the "GSH_" prefix are considered GSH-specific.
func IsGSHVariable(name string) bool {
	gshVars := []string{
		"GSH_PROMPT", "GSH_APROMPT", "GSH_LOG_LEVEL", "GSH_CLEAN_LOG_FILE",
		"GSH_MINIMUM_HEIGHT", "GSH_FAST_MODEL_API_KEY", "GSH_FAST_MODEL_BASE_URL",
		"GSH_FAST_MODEL_ID", "GSH_SLOW_MODEL_API_KEY", "GSH_SLOW_MODEL_BASE_URL",
		"GSH_SLOW_MODEL_ID", "GSH_CONTEXT_TYPES_FOR_AGENT", "GSH_AGENT_CONTEXT_WINDOW_TOKENS",
		"GSH_AGENT_APPROVED_BASH_COMMAND_REGEX", "GSH_AGENT_MACROS",
	}

	for _, gshVar := range gshVars {
		if name == gshVar {
			return true
		}
	}
	return strings.HasPrefix(name, "GSH_")
}