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
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/prometheus/client_golang/prometheus"
)

// standard service metric labels
const (
	// METRIC_LABEL_SERVICE svc="ServiceInterface" - used to group metrics by service
	// the label value format is : serviceInterface.PkgPath()/serviceInterface.Name()
	METRIC_LABEL_SERVICE = "svc"
	// METRIC_LABEL_SERVICE_VERSION svc-ver="X.Y.Z" - used to group service metrics by service version
	METRIC_LABEL_SERVICE_VERSION = "svc_ver"
)

// AddServiceMetricLabels adds labels to help identify service metrics. The above constants define the labels that are added, e.g.,
//
//		svc="github.com/oysterpack/oysterpack.go/cmd/demos/services/counter/Service",svc_ver="1.0.0"
//
func AddServiceMetricLabels(labels prometheus.Labels, serviceInterface ServiceInterface, version *semver.Version) prometheus.Labels {
	labels[METRIC_LABEL_SERVICE] = fmt.Sprintf("%v/%v", serviceInterface.PkgPath(), serviceInterface.Name())
	labels[METRIC_LABEL_SERVICE_VERSION] = version.String()
	return labels
}
