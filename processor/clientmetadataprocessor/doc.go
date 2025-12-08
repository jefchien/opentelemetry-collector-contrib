// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

// Package clientmetadataprocessor contains the logic to modify client.Metadata.
// It supports insert, update, upsert and delete as actions.
package clientmetadataprocessor // import "github.com/jefchien/opentelemetry-collector-contrib/processor/clientmetadataprocessor"
