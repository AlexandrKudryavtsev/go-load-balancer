package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	duration, err := time.ParseDuration(node.Value)
	if err != nil {
		return err
	}

	d.Duration = duration

	return nil
}
