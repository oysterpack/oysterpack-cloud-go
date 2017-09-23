// Copyright (c) 2017 OysterPack, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics_test

import (
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCounterFQName(t *testing.T) {
	opts := &prometheus.CounterOpts{
		Namespace:   "Namespace",
		Subsystem:   "Subsystem",
		Name:        "Name",
		ConstLabels: map[string]string{"a": "b"},
	}

	expected := "Namespace_Subsystem_Name"
	if actual := metrics.CounterFQName(opts); actual != expected {
		t.Errorf("%v != %v", actual, expected)
	}
}

func TestCounterOptsMatch(t *testing.T) {
	opts1 := &prometheus.CounterOpts{
		Namespace:   "Namespace",
		Subsystem:   "Subsystem",
		Name:        "Name",
		ConstLabels: map[string]string{"a": "b"},
	}

	opts2 := &prometheus.CounterOpts{
		Namespace:   "Namespace",
		Subsystem:   "Subsystem",
		Name:        "Name",
		ConstLabels: map[string]string{"a": "b"},
	}

	if !metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should have matched")
	}
	opts1.Namespace = ""
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
	opts1.Namespace = "Namespace"
	opts1.Subsystem = ""
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
	opts1.Namespace = "Namespace"
	opts1.Subsystem = "Subsystem"
	opts1.Name = ""
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts1.Namespace = "Namespace"
	opts1.Subsystem = "Subsystem"
	opts1.Name = "Name"
	opts1.ConstLabels = nil
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}

	opts1.Namespace = "Namespace"
	opts1.Subsystem = "Subsystem"
	opts1.Name = "Name"
	opts1.ConstLabels = opts2.ConstLabels
	opts1.Help = "HELP"
	if metrics.CounterOptsMatch(opts1, opts2) {
		t.Error("opts should not have matched")
	}
}
