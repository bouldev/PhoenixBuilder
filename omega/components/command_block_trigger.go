package components

import "phoenixbuilder/omega/defines"

type CBTrigger struct {
	*defines.BasicComponent
	TriggerPatterns        []string
	CompiledTriggerChecker func(input string) (keys map[string]string, hit bool)
}
