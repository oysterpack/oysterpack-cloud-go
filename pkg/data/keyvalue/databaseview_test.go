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

package keyvalue_test

import (
	"os"
	"testing"

	"github.com/oysterpack/oysterpack.go/pkg/data/keyvalue"
)

func TestOpenDatabaseView(t *testing.T) {
	dbFile := dbFile("TestOpenDatabaseView")
	deleteDbFile(t, dbFile)
	if err := os.Remove(dbFile); err != nil {
		t.Log(err)
	}
	dbName := "test"

	_, err := keyvalue.OpenDatabase(dbFile, dbName)
	if err == nil {
		t.Errorf("An error should have been returned because the database file does not exist")
	} else {
		t.Log(err)
	}

	// create the database file
	db, err := keyvalue.CreateDatabase(dbFile, dbName, true)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	dbView, err := keyvalue.OpenDatabaseView(dbFile, dbName)
	if err != nil {
		t.Fatal(err)
	}
	defer dbView.Close()
	if dbView.Name() != dbName {
		t.Errorf("db name does not match : %s", db.Name())
	}
}
