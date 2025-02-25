package pm

import (
	"fmt"
	"strings"
)

// PostmanContext stores additional Postman configuration
type PostmanContext struct {
	// Original auth handling
	Auth map[string][]*PostmanAuthParam
	// Environment variables from external file
	Environment map[string]string
	// Scripts storage (replaces exec array)
	Scripts map[string][]string
	// Collection variables
	CollectionVars map[string]string
	// Execution settings
	Settings    *PostmanSettings
	ScriptExecs []string
}

// PostmanSettings contains execution configuration
type PostmanSettings struct {
	TimeoutMS   int64
	KeepHeaders bool
	DelayMS     int64
}

// NewPostmanContext creates a new context
func NewPostmanContext() *PostmanContext {
	return &PostmanContext{
		Auth:           make(map[string][]*PostmanAuthParam),
		Environment:    make(map[string]string),
		Scripts:        make(map[string][]string),
		CollectionVars: make(map[string]string),
		ScriptExecs:    make([]string, 0),
		Settings: &PostmanSettings{
			TimeoutMS:   3000,
			KeepHeaders: true,
		},
	}
}

// AddAuth Auth helper methods
func (pc *PostmanContext) AddAuth(authType string, params []*PostmanAuthParam) {
	pc.Auth[authType] = params
}

func (pc *PostmanContext) GetAuth(authType string) []*PostmanAuthParam {
	return pc.Auth[authType]
}

// AddPreRequestScript Script helper methods
func (pc *PostmanContext) AddPreRequestScript(name, script string) {
	pc.AddScript(name, "prerequest", script)
}

func (pc *PostmanContext) AddTestScript(name, script string) {
	pc.AddScript(name, "test", script)
}

// AddScript helper methods
func (pc *PostmanContext) AddScript(name string, scriptType string, code string) {
	key := fmt.Sprintf("%s:%s", scriptType, name)
	pc.Scripts[key] = append(pc.Scripts[key], code)

	if scriptType == "prerequest" || scriptType == "test" {
		pc.ScriptExecs = append(pc.ScriptExecs, code)
	}
}

// GetScripts returns scripts of a specific type
func (pc *PostmanContext) GetScripts(scriptType string) []string {
	var scripts []string
	prefix := scriptType + ":"
	for key, values := range pc.Scripts {
		if strings.HasPrefix(key, prefix) {
			scripts = append(scripts, values...)
		}
	}
	return scripts
}

// SetVariable helper methods
func (pc *PostmanContext) SetVariable(scope string, name string, value string) {
	switch scope {
	case "collection":
		pc.CollectionVars[name] = value
	case "environment":
		pc.Environment[name] = value
	}
}

// GetVariable returns variable value with scope precedence
func (pc *PostmanContext) GetVariable(name string) string {
	// Environment overrides Collection
	if val, ok := pc.Environment[name]; ok {
		return val
	}
	return pc.CollectionVars[name]
}
