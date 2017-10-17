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
	"os"
	"strings"
	"time"

	"github.com/coreos/bbolt"
)

const (
	READ_MODE       os.FileMode = 0400
	READ_WRITE_MODE os.FileMode = 0600

	CREATED = "created"
)

// Database provides a read-write view of the database
type Database interface {
	Bucket

	// Created returns when the database was created, or an error if it cannot be determined
	Created() (time.Time, error)

	// Close releases all database resources. All transactions must be closed before closing the database.
	Close() error
}

// OpenDatabase opens the database in read-write mode
// The filePath must point to a bbolt file. The file must have the following structure :
// - a root bucket must exist that matches the database name
func OpenDatabase(filePath string, dbName string) (Database, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil, ErrFilePathIsBlank
	}

	if stat, err := os.Stat(filePath); err != nil {
		return nil, err
	} else if stat.IsDir() {
		return nil, errDatabaseFilePathIsDir(filePath)
	}

	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return nil, ErrDatabaseNameMustNotBeBlank
	}

	dbKey := []byte(dbName)

	options := &bolt.Options{
		Timeout: time.Second * 30,
	}

	db, err := bolt.Open(filePath, READ_WRITE_MODE, options)
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		dbBucket := tx.Bucket(dbKey)
		if dbBucket == nil {
			return errRootDatabaseBucketDoesNotExist(dbName)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	dbBucket := &bucket{&bucketView{name: string(dbName), path: []string{dbName}, db: db}}
	return &database{dbBucket}, nil
}

// CreateDatabase creates a new database with the specified name at the specified path
// ifNotExists = true -> if the database already exists then no error is returned
// ifNotExists = true -> if the database already exists, then an error is returned
//
// If the database file does not exist, then it will be created. The database existence is determined by the existence
// of a root bucket that matches the db name.
func CreateDatabase(filePath string, dbName string, ifNotExists bool) (Database, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil, ErrFilePathIsBlank
	}

	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return nil, ErrDatabaseNameMustNotBeBlank
	}

	options := &bolt.Options{
		Timeout: time.Second * 30,
	}

	db, err := bolt.Open(filePath, READ_WRITE_MODE, options)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if ifNotExists {
			if _, err = tx.CreateBucketIfNotExists([]byte(dbName)); err != nil {
				return err
			}
		} else {
			if tx.Bucket([]byte(dbName)) != nil {
				return ErrDatabaseBucketAlreadyExists
			}

			if _, err = tx.CreateBucket([]byte(dbName)); err != nil {
				return err
			}
		}

		// set the created timestamp
		dbBucket := tx.Bucket([]byte(dbName))
		if dbBucket.Get([]byte(CREATED)) == nil {
			now, _ := time.Now().MarshalBinary() // ignoring err, because this will never err
			dbBucket.Put([]byte(CREATED), now)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	dbBucket := &bucket{&bucketView{name: string(dbName), path: []string{dbName}, db: db}}
	return &database{dbBucket}, nil
}

type database struct {
	*bucket
}

func (a *database) Close() error {
	return a.db.Close()
}

func (a *database) Created() (time.Time, error) {
	t := time.Time{}
	err := t.UnmarshalBinary(a.Get(CREATED))
	return t, err
}
