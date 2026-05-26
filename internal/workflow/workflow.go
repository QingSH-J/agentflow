package workflow

// ParseWorkflowDefinition parses the workflow definition and returns a Workflow struct.
import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Definition struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	ID        string   `yaml:"id"`
	Type      string   `yaml:"type"`
	DependsOn []string `yaml:"depends_on"`
	Prompt    string   `yaml:"prompt"`
	Retry     int      `yaml:"retry"`
	Timeout   int      `yaml:"timeout"`
}

func LoadFile(path string) (*Definition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, err
	}

	if err := def.Validate(); err != nil {
		return nil, err
	}

	return &def, nil
}

func (d *Definition) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(d.Steps) == 0 {
		return fmt.Errorf("at least one step is required")
	}

	stepIDs := map[string]bool{}
	stepByID := map[string]Step{}
	for _, s := range d.Steps {
		if s.ID == "" {
			return fmt.Errorf("step id is required")
		}
		if s.Type == "" {
			return fmt.Errorf("step type is required")
		}
		if stepIDs[s.ID] {
			return fmt.Errorf("duplicate step id: %s", s.ID)
		}
		stepIDs[s.ID] = true
		stepByID[s.ID] = s
	}

	for _, s := range d.Steps {
		for _, dep := range s.DependsOn {
			if !stepIDs[dep] {
				return fmt.Errorf("step %s depends on unknown step %s", s.ID, dep)
			}
		}
	}

	//check for circular dependencies
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(stepID string) error
	visit = func(stepID string) error {
		if visiting[stepID] {
			return fmt.Errorf("circular dependency detected at step %s", stepID)
		}
		if visited[stepID] {
			return nil
		}
		visiting[stepID] = true
		step := stepByID[stepID]
		for _, dep := range step.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[stepID] = false
		visited[stepID] = true
		return nil
	}

	for _, s := range d.Steps {
		if err := visit(s.ID); err != nil {
			return err
		}
	}

	return nil
}
