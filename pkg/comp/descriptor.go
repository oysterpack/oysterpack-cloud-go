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

package comp

import (
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

// Descriptor is used to describe the service
// Think of the service as a component that is part of a system which belongs to a namespace.
// The service functionality is defined by its Interface.
// The service is versioned.
type Descriptor struct {
	namespace string
	system    string
	component string
	version   *semver.Version
}

// NewDescriptor creates a new descriptor.
// namespace, system, and component must not be blank, and must only consist of word characters.
// They will be trimmed and lower cased.
func NewDescriptor(namespace string, system string, component string, version string) *Descriptor {
	re, err := regexp.Compile("[[:word:]]")
	if err != nil {
		logger.Panic().Err(err).Msg("")
	}
	validate := func(name, s string) string {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			logger.Panic().Msgf("%q cannot be blank", name)
		}
		if !re.MatchString(s) {
			logger.Panic().Msgf("%q contains a non-word character : [%s]", name, s)
		}
		return strings.ToLower(s)
	}

	return &Descriptor{
		namespace: validate("namespace", namespace),
		system:    validate("system", system),
		component: validate("component", component),
		version:   NewVersion(version),
	}
}

// NewVersion returns a new version.
// If the version is not a valid semver, then the func panics.
func NewVersion(version string) *semver.Version {
	v, err := semver.NewVersion(version)
	if err != nil {
		logger.Panic().Msgf("Invalid version : %v : %v", version, err)
	}
	return v
}

// ID return the unique service id composed of its {namespace}-{system}-{component}-{version}
func (a *Descriptor) ID() string {
	return strings.Join([]string{a.namespace, a.system, a.component, a.version.String()}, "-")
}

func (a *Descriptor) String() string {
	return a.ID()
}

// Namespace returns the namespace that the service belongs to
func (a *Descriptor) Namespace() string {
	return a.namespace
}

// System returns the name of the system that the service belongs to
func (a *Descriptor) System() string {
	return a.system
}

// Component returns the name of the component
func (a *Descriptor) Component() string {
	return a.component
}

// Version returns the service version
func (a *Descriptor) Version() *semver.Version {
	return a.version
}
