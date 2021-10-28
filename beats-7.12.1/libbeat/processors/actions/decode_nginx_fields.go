// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package actions

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strings"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/libbeat/processors/checks"
	jsprocessor "github.com/elastic/beats/v7/libbeat/processors/script/javascript/module/processor"
)

type decodeNGINXFields struct {
	fields []string
	logkey string
	target *string
	logger *logp.Logger
}

type configNg struct {
	Fields []string `config:"fields"`
	LogKey string   `config:"logkey"`
	Target *string  `config:"target"`
}

var (
	defaultConfigNg = configNg{
		Fields: []string{"message"},
		LogKey: "[access]",
	}
	errProcessingSkippedNg = errors.New("processing skipped")
)

func init() {
	processors.RegisterPlugin("decode_nginx_fields",
		checks.ConfigChecked(NewDecodeNGINXFields,
			checks.RequireFields("fields"),
			checks.AllowedFields("fields", "logkey", "target")))

	jsprocessor.RegisterPlugin("DecodeNGINXFields", NewDecodeNGINXFields)
}

// NewDecodeNGINXFields construct a new decode_nginx_fields processor.
func NewDecodeNGINXFields(c *common.Config) (processors.Processor, error) {
	config := defaultConfigNg
	logger := logp.NewLogger("truncate_fields")

	err := c.Unpack(&config)
	if err != nil {
		logger.Warn("Error unpacking config for decode_nginx_fields")
		return nil, fmt.Errorf("fail to unpack the decode_nginx_fields configuration: %s", err)
	}

	f := &decodeNGINXFields{
		fields: config.Fields,
		logkey: config.LogKey,
		target: config.Target,
		logger: logger,
	}
	return f, nil
}

func (f *decodeNGINXFields) Run(event *beat.Event) (*beat.Event, error) {
	var errs []string

	for _, field := range f.fields {
		data, err := event.GetValue(field)
		if err != nil && errors.Cause(err) != common.ErrKeyNotFound {
			f.logger.Debugf("Error trying to GetValue for field : %s in event : %v", field, event)
			errs = append(errs, err.Error())
			continue
		}

		text, ok := data.(string)
		if !ok {
			// ignore non string fields when unmarshaling
			continue
		}

		output, err := decodeNGINX(text, f.logkey)
		if err != nil {
			f.logger.Debugf("Error trying to decodeNGINX %s", text)
			errs = append(errs, err.Error())
			continue
		}

		target := field
		if f.target != nil {
			target = *f.target
		}

		if target != "" {
			_, err = event.PutValue(target, output)
		}

		if err != nil {
			f.logger.Debugf("Error trying to Put value %v for field : %s", output, field)
			errs = append(errs, err.Error())
			continue
		}
	}

	if len(errs) > 0 {
		return event, fmt.Errorf(strings.Join(errs, ", "))
	}
	return event, nil
}

func decodeNGINX(text string, logkey string) (map[string]interface{}, error) {
	if strings.HasPrefix(text, logkey) {
		var parts [][]string
		var to map[string]interface{}
		matcher := regexp.MustCompile("(.+?)\\s")
		parts = matcher.FindAllStringSubmatch(text, -1)
		if len(parts) > 11 {
			to["remote_addr"] = parts[1][1]
			to["time_local"] = parts[4][1] + parts[5][1]
			if strings.Contains(parts[6][1], "\"") {
				method := parts[6][1]
				to["method"] = strings.Replace(method, "\"", "", 2)
			}
			to["path"] = parts[7][1]
			to["status"] = parts[9][1]
			to["body_bytes_sent"] = parts[10][1]
			to["http_referer"] = parts[11][1]
			to["http_user_agent"] = parts[12][1]

			return to, nil
		}
	}
	return nil, errProcessingSkippedNg
}

func (f decodeNGINXFields) String() string {
	return "decode_nginx_fields=" + strings.Join(f.fields, ", ")
}
