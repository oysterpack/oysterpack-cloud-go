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
	"fmt"
	"strings"
	"time"

	"github.com/coreos/bbolt"
)

// DatabaseView provides a read-only view of the database
type DatabaseView interface {
	Name() string

	// Buckets return the bucket names on the returned channel. Only Top-level buckets are returned.
	// The cancel channel is used to terminate the iteration early by the client.
	BucketViews(cancel <-chan struct{}) <-chan BucketView

	// Bucket returns the bucket for the specified name. If the bucket does not exist, then nil is returned.
	// If path is specified, then the database will traverse the path to locate the Bucket within its hierarchy.
	BucketView(path ...string) BucketView

	// Close releases all database resources. All transactions must be closed before closing the database.
	Close() error
}

// OpenDatabaseView opens the database in read-only mode. No other process may open it in read-write mode.
// Read-only mode uses a shared lock to allow multiple processes to read from the database but it will block any processes
// from opening the database in read-write mode.
//
// The path must point to a bbolt file. The file must have the following structure :
// - a root bucket must exist named 'db'
// - the root 'db' bucket must contain a key named 'name', which is the database name
func OpenDatabaseView(path string) (DatabaseView, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrPathIsRequired
	}

	options := &bolt.Options{
		ReadOnly: true,
		Timeout:  time.Second * 30,
	}

	db, err := bolt.Open(path, READ_MODE, options)
	if err != nil {
		return nil, err
	}

	var dbName []byte
	err = db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket(database_bucket)
		if root != nil {
			dbName = root.Get(database_name)
		} else {
			err = fmt.Errorf("Root database bucket does not exist : %s", DATABASE_BUCKET)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(dbName) == 0 {
		return nil, ErrDatabaseNameIsRequired
	}
	dbBucket := &bucketView{name: string(dbName), path: database_path, db: db}
	return &databaseView{dbBucket}, nil
}

type databaseView struct {
	*bucketView
}

func (a *databaseView) Name() string {
	return string(a.Get(DATABASE_NAME))
}

func (a *databaseView) Close() error {
	return a.db.Close()
}
