package shared

import (
	"encoding/json"
	"fmt"
	"io"
)

// Builder handles construction of hook output JSON to stdout
type Builder struct {
	writer io.Writer
}

// NewBuilder creates a new Builder instance
func NewBuilder(writer io.Writer) *Builder {
	return &Builder{writer: writer}
}

// BuildAllow builds a simple "allow" response for the specified hook event
func (b *Builder) BuildAllow(hookEventName string) error {
	output := HookOutput{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName:      hookEventName,
			PermissionDecision: "allow",
		},
	}
	return b.write(output)
}

// BuildAllowWithReason builds an "allow" response with a reason
func (b *Builder) BuildAllowWithReason(hookEventName, reason string) error {
	output := HookOutput{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName:            hookEventName,
			PermissionDecision:       "allow",
			PermissionDecisionReason: reason,
		},
	}
	return b.write(output)
}

// BuildDeny builds a "deny" response with a reason
func (b *Builder) BuildDeny(hookEventName, reason string) error {
	output := HookOutput{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName:            hookEventName,
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		},
	}
	return b.write(output)
}

// BuildWithUpdatedInput builds a response with updated tool input
func (b *Builder) BuildWithUpdatedInput(hookEventName string, updatedInput map[string]interface{}) error {
	output := HookOutput{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName:      hookEventName,
			PermissionDecision: "allow",
			UpdatedInput:       updatedInput,
		},
	}
	return b.write(output)
}

// BuildEmpty builds an empty response (for notification hooks that don't return data)
func (b *Builder) BuildEmpty(hookEventName string) error {
	output := HookOutput{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName: hookEventName,
		},
	}
	return b.write(output)
}

// BuildCustom builds a response with custom hook-specific output
func (b *Builder) BuildCustom(output HookOutput) error {
	return b.write(output)
}

// write encodes the output as JSON and writes it to the writer
func (b *Builder) write(output HookOutput) error {
	encoder := json.NewEncoder(b.writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode hook output: %w", err)
	}
	return nil
}
