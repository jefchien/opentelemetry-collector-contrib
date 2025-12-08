// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package clientmetadataprocessor // import "github.com/jefchien/opentelemetry-collector-contrib/processor/clientmetadataprocessor"

import (
	"errors"
	"fmt"

	"github.com/jefchien/opentelemetry-collector-contrib/processor/clientmetadataprocessor/internal/actions"
	"go.opentelemetry.io/collector/component"
)

// Config defines the configuration for the clientmetadata processor.
type Config struct {
	// Actions specifies the list of actions to perform on client metadata.
	// This is a required field.
	Actions []actions.KeyValue `mapstructure:"actions"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.Actions) == 0 {
		return errors.New("missing required field \"actions\"")
	}

	for i, action := range cfg.Actions {
		if action.Key == "" {
			return fmt.Errorf("action %d: missing required field \"key\"", i)
		}
		if action.Action == "" {
			return fmt.Errorf("action %d: missing required field \"action\"", i)
		}

		sourceCount := 0
		if action.Value != nil {
			sourceCount++
		}
		if action.FromAttribute != "" {
			sourceCount++
		}
		if action.FromResourceAttribute != "" {
			sourceCount++
		}

		if action.Action == actions.DELETE {
			if sourceCount != 0 {
				return fmt.Errorf("action %d: DELETE action should not specify value sources", i)
			}
		} else {
			if sourceCount != 1 {
				return fmt.Errorf("action %d: exactly one of \"value\", \"from_attribute\", or \"from_resource_attribute\" must be specified", i)
			}
		}
	}

	return nil
}
