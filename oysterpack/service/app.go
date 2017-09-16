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

var app Application = NewApplication(ApplicationSettings{})

// App exposes the Application globally.
//
// Use cases:
// 1. Package init functions use it to to register services when the package is loaded
// 2. Used to register services in the main function
// 3. Used to integrate application services with third party libraries.
func App() Application { return app }
