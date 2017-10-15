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

package keyvalue

// DatabaseRegistry is a database registry
type DatabaseRegistry interface {
	// DatabaseNames returns names of databases that the registry is aware of
	DatabaseNames() []string

	// OpenDatabaseNames returns the names of databases that are currently open
	OpenDatabaseNames() []string

	// DatabaseView will return the specified database if it is registered.
	// NOTE: a database can only be opened in read-only mode or read-write mode, i.e., it is not allowed to open the database in both modes
	DatabaseView(name string) (DatabaseView, error)

	// Database will return the specified database if it is registered.
	// NOTE: a database can only be opened in read-only mode or read-write mode, i.e., it is not allowed to open the database in both modes
	Database(name string) (Database, error)

	// CloseAll closes all open databases
	CloseAll()
}
