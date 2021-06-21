// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by beats/dev-tools/cmd/asset/asset.go - DO NOT EDIT.

package gsuite

import (
	"github.com/elastic/beats/v7/libbeat/asset"
)

func init() {
	if err := asset.SetFields("filebeat", "gsuite", asset.ModuleFieldsPri, AssetGsuite); err != nil {
		panic(err)
	}
}

// AssetGsuite returns asset data.
// This is the base64 encoded gzipped contents of module/gsuite.
func AssetGsuite() string {
	return "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"
}
