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
	"strings"
	"time"

	"os"

	"github.com/coreos/bbolt"
)

// DatabaseView provides a read-only view of the database
type DatabaseView interface {
	BucketView

	// Created returns when the database was created, or an error if it cannot be determined
	Created() (time.Time, error)

	// Close releases all database resources. All transactions must be closed before closing the database.
	Close() error
}

// OpenDatabaseView opens the database in read-only mode. No other process may open it in read-write mode.
// Read-only mode uses a shared lock to allow multiple processes to read from the database but it will block any processes
// from opening the database in read-write mode.
//
// The path must point to a bbolt file. The file must have the following structure :
// - a root bucket must exist with the db name
// - the root 'db' bucket must contain a key named 'name', which is the database name
func OpenDatabaseView(filePath string, dbName string) (DatabaseView, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil, ErrFilePathIsBlank
	}

	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return nil, ErrDatabaseNameMustNotBeBlank
	}

	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}

	options := &bolt.Options{
		ReadOnly: true,
		Timeout:  time.Second * 30,
	}

	db, err := bolt.Open(filePath, READ_MODE, options)
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		dbBucket := tx.Bucket([]byte(dbName))
		if dbBucket == nil {
			return errRootDatabaseBucketDoesNotExist(dbName)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dbBucket := &bucketView{name: string(dbName), path: []string{dbName}, db: db}
	return &databaseView{dbBucket}, nil
}

type databaseView struct {
	*bucketView
}

func (a *databaseView) Close() error {
	return a.db.Close()
}

func (a *databaseView) Created() (time.Time, error) {
	t := time.Time{}
	err := t.UnmarshalBinary(a.Get(CREATED))
	return t, err
}
