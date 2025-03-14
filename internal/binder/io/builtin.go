// Copyright 2021 EMQ Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package io

import (
	"github.com/lf-edge/ekuiper/internal/topo/sink"
	"github.com/lf-edge/ekuiper/internal/topo/source"
	"github.com/lf-edge/ekuiper/pkg/api"
)

type NewSourceFunc func() api.Source
type NewSinkFunc func() api.Sink

var (
	sources = map[string]NewSourceFunc{
		"mqtt":     func() api.Source { return &source.MQTTSource{} },
		"httppull": func() api.Source { return &source.HTTPPullSource{} },
		"file":     func() api.Source { return &source.FileSource{} },
	}
	sinks = map[string]NewSinkFunc{
		"log":         sink.NewLogSink,
		"logToMemory": sink.NewLogSinkToMemory,
		"mqtt":        func() api.Sink { return &sink.MQTTSink{} },
		"rest":        func() api.Sink { return &sink.RestSink{} },
		"nop":         func() api.Sink { return &sink.NopSink{} },
	}
)

type Manager struct{}

func (m *Manager) Source(name string) (api.Source, error) {
	if s, ok := sources[name]; ok {
		return s(), nil
	}
	return nil, nil
}

func (m *Manager) Sink(name string) (api.Sink, error) {
	if s, ok := sinks[name]; ok {
		return s(), nil
	}
	return nil, nil
}

var m = &Manager{}

func GetManager() *Manager {
	return m
}
