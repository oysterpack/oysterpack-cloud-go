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

import (
	"errors"
	"fmt"
)

var (
	ErrFilePathIsBlank             = errors.New("Path must not be blank")
	ErrDatabaseNameMustNotBeBlank  = errors.New("Database name must not be blank.")
	ErrDatabaseBucketAlreadyExists = errors.New("Database bucket already exists")
)

func errBucketDoesNotExist(path []string) error {
	return fmt.Errorf("Bucket does not exist at path : %v", path)
}

func errRootDatabaseBucketDoesNotExist(dbName string) error {
	return fmt.Errorf("Root database bucket does not exist : %s", dbName)
}

func errDatabaseFilePathIsDir(filePath string) error {
	return fmt.Errorf("The database file path must point to a file, not a directory : %s", filePath)
}
