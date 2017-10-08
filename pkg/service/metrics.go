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

package service

import (
	"github.com/prometheus/client_golang/prometheus"
)

// standard service metric labels
const (
	// METRIC_LABEL_NAMESPACE ns -> Descriptor.NameSpace
	METRIC_LABEL_NAMESPACE = "ns"
	// METRIC_LABEL_SYSTEM  sys -> Descriptor.System
	METRIC_LABEL_SYSTEM = "sys"
	// METRIC_LABEL_COMPONENT comp -> Descriptor.Component
	METRIC_LABEL_COMPONENT = "comp"
	// METRIC_LABEL_VERSION ver -> Descriptor.Version
	METRIC_LABEL_VERSION = "ver"
)

// AddServiceMetricLabels adds labels to help identify service metrics. The above constants define the labels that are added, e.g.,
//
//		ns="oysterpack",sys="metrics",comp="http",ver="1.0.0"
//
func AddServiceMetricLabels(labels prometheus.Labels, desc *Descriptor) prometheus.Labels {
	labels[METRIC_LABEL_NAMESPACE] = desc.Namespace()
	labels[METRIC_LABEL_SYSTEM] = desc.System()
	labels[METRIC_LABEL_COMPONENT] = desc.Component()
	labels[METRIC_LABEL_VERSION] = desc.Version().String()
	return labels
}
