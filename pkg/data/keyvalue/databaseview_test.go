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

	"time"

	"github.com/nats-io/nuid"
	"github.com/oysterpack/oysterpack.go/pkg/data/keyvalue"
)

func TestOpenDatabaseView(t *testing.T) {
	startOfTest := time.Now()
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

	created, err := dbView.Created()
	if err != nil {
		t.Errorf("Failed to get the database create timestamp : %v", err)
	}
	t.Logf("Database created on : %v", created)
	if !created.After(startOfTest) {
		t.Errorf("The created time should be after the start of the test : %v <= %v", created, startOfTest)
	}
}

func TestOpenDatabase_BlankFilePath(t *testing.T) {
	if _, err := keyvalue.OpenDatabase(" ", "test"); err == nil {
		t.Errorf("Should have failed because the file path is blank")
	} else {
		t.Log(err)
	}

	if _, err := keyvalue.OpenDatabaseView(" ", "test"); err == nil {
		t.Errorf("Should have failed because the file path is blank")
	} else {
		t.Log(err)
	}
}

func TestOpenDatabase_BlankDBName(t *testing.T) {
	if _, err := keyvalue.OpenDatabase("asdasd", "  "); err == nil {
		t.Errorf("Should have failed because the db name is blank")
	} else {
		t.Log(err)
	}

	if _, err := keyvalue.OpenDatabaseView("asdasd", "  "); err == nil {
		t.Errorf("Should have failed because the db name is blank")
	} else {
		t.Log(err)
	}
}

func TestOpenDatabase_DBFileDoesNotExist(t *testing.T) {
	if _, err := keyvalue.OpenDatabase(nuid.Next(), "test"); err == nil {
		t.Errorf("Should have failed because the db file does not exist")
	} else {
		t.Log(err)
	}

	if _, err := keyvalue.OpenDatabaseView(nuid.Next(), "test"); err == nil {
		t.Errorf("Should have failed because the db file does not exist")
	} else {
		t.Log(err)
	}
}

func TestOpenDatabase_DBFileIsDir(t *testing.T) {
	if _, err := keyvalue.OpenDatabase("./testdata", "test"); err == nil {
		t.Errorf("Should have failed because the file path points to a dir")
	} else {
		t.Log(err)
	}

	if _, err := keyvalue.OpenDatabaseView("./testdata", "test"); err == nil {
		t.Errorf("Should have failed because the file path points to a dir")
	} else {
		t.Log(err)
	}
}

func TestOpenDatabase_DBBucketIsMissing(t *testing.T) {
	dbFile := dbFile("TestOpenDatabaseView")
	deleteDbFile(t, dbFile)
	dbName := "test"
	db, err := keyvalue.CreateDatabase(dbFile, dbName, true)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	if _, err := keyvalue.OpenDatabase(dbFile, nuid.Next()); err == nil {
		t.Errorf("Should have failed because the the db bucket does not exist")
	} else {
		t.Log(err)
	}

	if _, err := keyvalue.OpenDatabaseView(dbFile, nuid.Next()); err == nil {
		t.Errorf("Should have failed because the the db bucket does not exist")
	} else {
		t.Log(err)
	}
}
