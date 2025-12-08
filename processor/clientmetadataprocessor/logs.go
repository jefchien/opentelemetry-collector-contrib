// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package clientmetadataprocessor // import "github.com/jefchien/opentelemetry-collector-contrib/processor/clientmetadataprocessor"

import (
	"context"

	"github.com/jefchien/opentelemetry-collector-contrib/processor/clientmetadataprocessor/internal/actions"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
)

type logsProcessor struct {
	actions actions.Actions
	next    consumer.Logs
}

func newLogsProcessor(cfg *Config, next consumer.Logs) processor.Logs {
	return &logsProcessor{
		actions: actions.NewActions(cfg.Actions),
		next:    next,
	}
}

func (p *logsProcessor) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	clientInfo := client.FromContext(ctx)
	metadataMap := make(map[string][]string)
	for key := range clientInfo.Metadata.Keys() {
		metadataMap[key] = clientInfo.Metadata.Get(key)
	}

	// Process static actions once
	p.actions.ProcessStatic(metadataMap)

	if p.actions.HasResourceActions() || p.actions.HasAttributeActions() {
		resourceLogs := ld.ResourceLogs()
		for i := 0; i < resourceLogs.Len(); i++ {
			rl := resourceLogs.At(i)

			// Process resource-level actions once per resource
			if p.actions.HasResourceActions() {
				p.actions.ProcessResource(metadataMap, rl.Resource().Attributes())
			}

			// Process attribute-level actions for each log record
			if p.actions.HasAttributeActions() {
				scopeLogs := rl.ScopeLogs()
				for j := 0; j < scopeLogs.Len(); j++ {
					logs := scopeLogs.At(j).LogRecords()
					for k := 0; k < logs.Len(); k++ {
						p.actions.ProcessAttributes(metadataMap, logs.At(k).Attributes())
					}
				}
			}
		}
	}

	clientInfo.Metadata = client.NewMetadata(metadataMap)
	newCtx := client.NewContext(ctx, clientInfo)
	return p.next.ConsumeLogs(newCtx, ld)
}

func (*logsProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (*logsProcessor) Start(context.Context, component.Host) error {
	return nil
}

func (*logsProcessor) Shutdown(context.Context) error {
	return nil
}
