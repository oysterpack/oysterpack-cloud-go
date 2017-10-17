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
	"testing"

	"os"

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/data/keyvalue"
)

const BASE_DB_PATH = "./testdata/temp/database_test_"

func dbFile(dbName string) string {
	return BASE_DB_PATH + dbName
}

func deleteDbFile(t *testing.T, dbName string) {
	t.Helper()
	if err := os.Remove(dbFile(dbName)); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestCreateDatabase(t *testing.T) {
	startOfTest := time.Now()
	dbFile := dbFile("TestCreateDatabase")
	deleteDbFile(t, dbFile)
	if err := os.Remove(dbFile); err != nil {
		t.Log(err)
	}
	dbName := "test"

	db, err := keyvalue.CreateDatabase(dbFile, dbName, true)
	if err != nil {
		t.Fatal(err)
	}

	fileStats, err := os.Stat(dbFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("name = %s, size = %d, mode = %v, modTime = %v, isDir = %v", fileStats.Name(), fileStats.Size(), fileStats.Mode(), fileStats.ModTime(), fileStats.IsDir())

	if db.Name() != dbName {
		t.Errorf("db name does not match : %s", db.Name())
	}
	created, err := db.Created()
	if err != nil {
		t.Errorf("Failed to get the database create timestamp : %v", err)
	}
	t.Logf("Database created on : %v", created)
	if !created.After(startOfTest) {
		t.Errorf("The created time should be after the start of the test : %v <= %v", created, startOfTest)
	}

	db.Close()

	// when ifNotExists=true, CreateDatabase will simply open the existing database
	db, err = keyvalue.CreateDatabase(dbFile, dbName, true)
	if err != nil {
		t.Fatal(err)
	}
	created2, err := db.Created()
	if !created2.Equal(created) {
		t.Errorf("The created timestamp should be the same as before : %v", created2)
	}
	db.Close()

	fileStats, err = os.Stat(dbFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("name = %s, size = %d, mode = %v, modTime = %v, isDir = %v", fileStats.Name(), fileStats.Size(), fileStats.Mode(), fileStats.ModTime(), fileStats.IsDir())

	db, err = keyvalue.CreateDatabase(dbFile, dbName, false)
	if err == nil {
		t.Error("An error should have been returned because the database already exists")
	} else {
		t.Log(err)
	}
}

func TestOpenDatabase(t *testing.T) {
	dbFile := dbFile("TestOpenDatabase")
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

	db, err = keyvalue.OpenDatabase(dbFile, dbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if db.Name() != dbName {
		t.Errorf("db name does not match : %s", db.Name())
	}
}
