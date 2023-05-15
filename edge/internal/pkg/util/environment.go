package util

import (
	"github.com/rotisserie/eris"
	"strings"
	"sync"
)

var (
	ErrUnmatchedLeftBrace  = eris.New("Unmatched left brace")
	ErrUnmatchedRightBrace = eris.New("Unmatched right brace")
	ErrVariableNotFound    = eris.New("Variable not found")
)

// Environment is a utility structure to handle hierarchies of string variables.
type Environment struct {
	mu            sync.RWMutex
	variables     map[string]string
	parent        *Environment
	caseSensitive bool
}

// NewEnvironment creates a new empty environment.
// If the environment shouldn't have a parent, set the parent argument
// to nil
func NewEnvironment(parent *Environment) *Environment {
	caseSensitive := false
	if parent != nil {
		caseSensitive = parent.caseSensitive
	}

	return &Environment{
		variables:     make(map[string]string),
		parent:        parent,
		caseSensitive: caseSensitive,
	}
}

func NewEnvironmentCaseSensitive(parent *Environment) *Environment {
	return &Environment{
		variables:     make(map[string]string),
		parent:        parent,
		caseSensitive: true,
	}
}

// Parent returns a pointer to parent environment
func (e *Environment) Parent() *Environment {
	return e.parent
}

func (e *Environment) IsCaseSensitive() bool {
	return e.caseSensitive
}

func (e *Environment) convName(name string) string {
	if e.caseSensitive {
		return name
	}

	return strings.ToLower(name)
}

// Set sets a variable in the current environment
func (e *Environment) Set(name, value string) *Environment {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.variables[e.convName(name)] = value

	return e
}

// SetFromMap sets multiple variables reading them from a map
func (e *Environment) SetFromMap(variablesMap map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for k, v := range variablesMap {
		e.variables[e.convName(k)] = v
	}
}

// Exists checks if a variable is defined in the current environment
// or in any parent
func (e *Environment) Exists(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if val, ok := e.variables[e.convName(name)]; ok && val != "" {
		return true
	}

	if e.parent != nil {
		return e.parent.Exists(name)
	}

	return false
}

// Get gets the value of a variable from the current environment or
// in any parent.
// If the variable is not set, an empty string is returned.
func (e *Environment) Get(name string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if val, ok := e.variables[e.convName(name)]; ok {
		return val
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return ""
}

// Clear removes all variables from the current environment.
// Parent variables won't be modified
func (e *Environment) Clear() *Environment {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.variables = make(map[string]string)
	return e
}

// ToMap builds a new map containing all variables declared
// through all parents
func (e *Environment) ToMap() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]string)

	if e.parent != nil {
		parentMap := e.parent.ToMap()
		for k, v := range parentMap {
			result[k] = v
		}
	}

	for k, v := range e.variables {
		result[k] = v
	}

	return result
}

// EscapeStringVariables function will replace all variable inside a given string.
// Variable values are read from the environment.
// Variable names must be enclosed inside braces, e.g.: "{variable1}"
func (e *Environment) EscapeStringVariables(str string) (string, error) {
	// Output string builder
	var outBuilder strings.Builder

	// Set capacity to original string length.
	// This optimization works best for strings without variables inside
	outBuilder.Grow(len(str))

	// State
	parsingVariableName := false
	variableName := ""
	lastLeftBracePos := 0

	for i, curChar := range str {
		if curChar == '{' {
			if parsingVariableName {
				return "", eris.Wrapf(ErrUnmatchedLeftBrace, "->{%s", str[i:])
			}
			parsingVariableName = true
			lastLeftBracePos = i
		} else if curChar == '}' {
			if !parsingVariableName {
				return "", eris.Wrapf(ErrUnmatchedRightBrace, "%s}<-", str[:i])
			}

			variableValue := e.Get(variableName)

			if variableValue != "" {
				outBuilder.WriteString(variableValue)
			} else {
				return "", eris.Wrap(ErrVariableNotFound, variableName)
			}

			variableName = ""
			parsingVariableName = false
		} else if parsingVariableName {
			variableName += string(curChar)
		} else {
			outBuilder.WriteRune(curChar)
		}
	}

	if parsingVariableName {
		// If we're still parsing a variable name after string end, it means
		// the last right brace is missing
		return "", eris.Wrapf(ErrUnmatchedLeftBrace, "->%s", str[lastLeftBracePos:])
	}

	return outBuilder.String(), nil
}
