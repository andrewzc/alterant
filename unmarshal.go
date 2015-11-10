package main

import (
	"os"
	"path"
)

func (m *machine) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux struct {
		Environment map[string]string `yaml:"environment"`
		Tasks       []string          `yaml:"tasks"`
	}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	m.Environment = aux.Environment

	m.Tasks = map[string]*task{}
	for _, taskName := range aux.Tasks {
		// temporary pointer to task type
		m.Tasks[taskName] = &task{}
	}

	return nil
}

func (c *config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux struct {
		Machines  map[string]*machine `yaml:"machines"`
		Tasks     map[string]*task    `yaml:"tasks"`
		Encrypted []string            `yaml:"encrypted"`
	}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	// match the machine's requested tasks to the tasks defined in the config
	c.actions = map[string]*OrderedMap{}
	for machineName, machinePtr := range aux.Machines {
		c.actions[machineName] = NewOrderedMap()
		for taskName, taskPtr := range aux.Tasks {
			_, ok := machinePtr.Tasks[taskName]
			if ok {
				taskPtr.name = taskName
				c.actions[machineName].Add(taskName, taskPtr)
			}
		}
	}

	c.encrypted = aux.Encrypted

	return nil
}

func (l *linkTarget) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux string

	if err := unmarshal(&aux); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	l.Value = path.Join(cwd, os.ExpandEnv(aux))

	return nil
}

func (l *linkDestination) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux string

	if err := unmarshal(&aux); err != nil {
		return err
	}

	l.Value = path.Join(os.Getenv("HOME"), os.ExpandEnv(aux))

	return nil
}
