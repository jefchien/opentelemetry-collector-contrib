// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package actions // import "github.com/jefchien/opentelemetry-collector-contrib/processor/clientmetadataprocessor/internal/actions"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// Action is the enum to capture the four types of actions to perform on client metadata.
type Action string

const (
	// INSERT adds the key/value to client metadata when the key does not exist.
	// No action is applied to client metadata where the key already exists.
	INSERT Action = "insert"

	// UPDATE updates an existing key with a value. No action is applied
	// to client metadata where the key does not exist.
	UPDATE Action = "update"

	// UPSERT performs the INSERT or UPDATE action. The key/value is
	// inserted to client metadata that did not originally have the key. The key/value is
	// updated for client metadata where the key already existed.
	UPSERT Action = "upsert"

	// DELETE deletes the key from the client metadata. If the key doesn't exist, no action is performed.
	DELETE Action = "delete"
)

// KeyValue specifies the client metadata key to act upon.
type KeyValue struct {
	// Key specifies the client metadata key to set.
	// This is a required field.
	Key string `mapstructure:"key"`

	// Action specifies the type of action to perform.
	// This is a required field.
	Action Action `mapstructure:"action"`

	// Value specifies the static value to set.
	Value any `mapstructure:"value"`

	// FromAttribute specifies the attribute to use to populate the value.
	// If the attribute doesn't exist, no action is performed.
	FromAttribute string `mapstructure:"from_attribute"`

	// FromResourceAttribute specifies the resource attribute to use to populate the value.
	// If the attribute doesn't exist, no action is performed.
	FromResourceAttribute string `mapstructure:"from_resource_attribute"`
}

type metadataAction struct {
	Key                   string
	Action                Action
	StaticValue           string
	FromResourceAttribute string
	FromAttribute         string
}

// Actions defines the interface for processing client metadata actions.
type Actions interface {
	ProcessStatic(metadata map[string][]string)
	ProcessResource(metadata map[string][]string, attrs pcommon.Map)
	ProcessAttributes(metadata map[string][]string, attrs pcommon.Map)
	HasResourceActions() bool
	HasAttributeActions() bool
}

type actions struct {
	staticActions    []metadataAction
	resourceActions  []metadataAction
	attributeActions []metadataAction
}

func NewActions(keyValues []KeyValue) Actions {
	var staticActions []metadataAction
	var resourceActions []metadataAction
	var attributeActions []metadataAction

	for _, kv := range keyValues {
		action := metadataAction{
			Key:    kv.Key,
			Action: kv.Action,
		}

		switch {
		case kv.Value != nil:
			action.StaticValue = kv.Value.(string)
			staticActions = append(staticActions, action)
		case kv.FromResourceAttribute != "":
			action.FromResourceAttribute = kv.FromResourceAttribute
			resourceActions = append(resourceActions, action)
		case kv.FromAttribute != "":
			action.FromAttribute = kv.FromAttribute
			attributeActions = append(attributeActions, action)
		case kv.Action == DELETE:
			staticActions = append(staticActions, action)
		}
	}

	return &actions{
		staticActions:    staticActions,
		resourceActions:  resourceActions,
		attributeActions: attributeActions,
	}
}

func (a *actions) HasResourceActions() bool {
	return len(a.resourceActions) > 0
}

func (a *actions) HasAttributeActions() bool {
	return len(a.attributeActions) > 0
}

func (a *actions) ProcessStatic(metadata map[string][]string) {
	for _, action := range a.staticActions {
		if action.StaticValue != "" {
			processAction(metadata, action, action.StaticValue)
		} else if action.Action == DELETE {
			processAction(metadata, action, "")
		}
	}
}

func (a *actions) ProcessResource(metadata map[string][]string, attrs pcommon.Map) {
	for _, action := range a.resourceActions {
		if value := getAttributeValue(attrs, action.FromResourceAttribute); value != "" {
			processAction(metadata, action, value)
		}
	}
}

func (a *actions) ProcessAttributes(metadata map[string][]string, attrs pcommon.Map) {
	for _, action := range a.attributeActions {
		if value := getAttributeValue(attrs, action.FromAttribute); value != "" {
			processAction(metadata, action, value)
		}
	}
}

func getAttributeValue(attrs pcommon.Map, key string) string {
	if key != "" {
		if val, found := attrs.Get(key); found {
			return val.AsString()
		}
	}
	return ""
}

func processAction(metadata map[string][]string, action metadataAction, value string) {
	switch action.Action {
	case INSERT:
		if _, exists := metadata[action.Key]; !exists {
			metadata[action.Key] = []string{value}
		}
	case UPDATE:
		if _, exists := metadata[action.Key]; exists {
			metadata[action.Key] = []string{value}
		}
	case UPSERT:
		metadata[action.Key] = []string{value}
	case DELETE:
		delete(metadata, action.Key)
	}
}
