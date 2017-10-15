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
	"github.com/coreos/bbolt"
)

// Bucket represents a bucket of key-value pairs. Keys are strings, but values are simply bytes.
// Buckets can form a hierarchy of buckets.
type Bucket interface {
	BucketView

	// Put sets the value for a key in the bucket. If the key exist then its previous value will be overwritten.
	// Supplied value must remain valid for the life of the transaction.
	// Returns an error if the key is blank, if the key is too large, or if the value is too large.
	Put(key string, value []byte) error

	// PutMultiple will put all values received on the data channel within the same transaction, i.e., either all or none will be stored.
	// Once nil is received on the data channel, then this signals to commit transaction.
	// The returned channel is to communicate an error, if the transaction failed to commit.
	PutMultiple(data <-chan *KeyValue) <-chan error

	// Delete removes the keys from the bucket. If the key does not exist then nothing is done and a nil error is returned.
	// All or none are deleted within the same transaction.
	Delete(key ...string) error

	// CreateBucket creates a new bucket at the given key and returns the new bucket.
	// Returns an error if the key already exists, if the bucket name is blank, or if the bucket name is too long.
	CreateBucket(name string) (Bucket, error)

	// CreateBucketIfNotExists creates a new bucket if it doesn't already exist and returns a reference to it.
	// Returns an error if the bucket name is blank, or if the bucket name is too long.
	CreateBucketIfNotExists(name string) (Bucket, error)

	// DeleteBucket deletes a bucket at the given key. Returns an error if the bucket does not exists, or if the key represents a non-bucket value.
	DeleteBucket(name string) error

	// Buckets iterate through the top-level children buckets and returns them on the returned channel.
	// The cancel channel is used to terminate the iteration early by the client.
	Buckets(cancel <-chan struct{}) <-chan Bucket

	// Bucket returns the bucket for the specified name. If the bucket does not exist, then nil is returned.
	// If path is specified, then the bucket will traverse the path to locate the Bucket within its hierarchy.
	Bucket(path ...string) Bucket
}

// BucketView is a read-only view of the Bucket
type BucketView interface {
	// Name returns the Bucket name
	Name() string

	// Get returns the value for the specified key
	// If the key does not exist, or if the key actually refers to a child Bucket, then nil is returned
	Get(key string) []byte

	// Keys returns the keys stored in this bucket. Keys are sorted, thus seek may be used to seek a position to start iterating.
	// seek is optional - if specified, then seek moves the cursor to a given key and returns it. If the key does not exist then the next key is used.
	Keys(seek string, cancel <-chan struct{}) <-chan string

	// KeyValues iterates through all key-value pairs and returns them via the channel.
	// The cancel channel is used to terminate the iteration early by the client.
	KeyValues(seek string, cancel <-chan struct{}) <-chan *KeyValue

	// BucketViews iterate through the top-level children buckets and returns them on the returned channel.
	// The cancel channel is used to terminate the iteration early by the client.
	BucketViews(cancel <-chan struct{}) <-chan BucketView

	// BucketView returns the bucket for the specified name. If the bucket does not exist, then nil is returned.
	// If path is specified, then the bucket will traverse the path to locate the Bucket within its hierarchy.
	BucketView(path ...string) BucketView
}

type KeyValue struct {
	Key   string
	Value []byte
}

type bucketView struct {
	name string

	path []string

	db *bolt.DB
}

func (a *bucketView) Name() string {
	return a.name
}

// Get returns the value for the specified key
// If the key does not exist, or if the key actually refers to a child Bucket, then nil is returned
func (a *bucketView) Get(key string) []byte {
	data := make(chan []byte, 1)

	a.db.View(func(tx *bolt.Tx) error {
		b := bucket(tx, nil, a.path)
		if b == nil {
			data <- nil
			return nil
		}

		data <- b.Get([]byte(key))
		return nil
	})

	return <-data
}

// Keys returns the keys stored in this bucket. Keys are sorted, thus seek may be used to seek a position to start iterating.
// seek is optional - if specified, then seek moves the cursor to a given key and returns it. If the key does not exist then the next key is used.
func (a *bucketView) Keys(seek string, cancel <-chan struct{}) <-chan string {
	c := make(chan chan string)

	go a.db.View(func(tx *bolt.Tx) error {
		data := make(chan string)
		c <- data
		b := bucket(tx, nil, a.path)
		if b == nil {
			close(data)
			return nil
		}

		cursor := b.Cursor()
		for k, _ := cursor.Seek([]byte(seek)); k != nil; k, _ = cursor.Next() {
			select {
			case <-cancel:
				break
			default:
				data <- string(k)
			}
		}
		close(data)

		return nil
	})

	return <-c
}

// KeyValues iterates through all key-value pairs and returns them via the channel.
// The cancel channel is used to terminate the iteration early by the client.
func (a *bucketView) KeyValues(seek string, cancel <-chan struct{}) <-chan *KeyValue {
	c := make(chan chan *KeyValue)

	go a.db.View(func(tx *bolt.Tx) error {
		data := make(chan *KeyValue)
		c <- data
		b := bucket(tx, nil, a.path)
		if b == nil {
			close(data)
			return nil
		}

		cursor := b.Cursor()
		for k, v := cursor.Seek([]byte(seek)); k != nil; k, v = cursor.Next() {
			select {
			case <-cancel:
				break
			default:
				data <- &KeyValue{string(k), v}
			}

		}
		close(data)

		return nil
	})

	return <-c
}

func (a *bucketView) BucketViews(cancel <-chan struct{}) <-chan BucketView {
	c := make(chan chan BucketView)

	go a.db.View(func(tx *bolt.Tx) error {
		buckets := make(chan BucketView)
		c <- buckets
		b := bucket(tx, nil, a.path)
		if b == nil {
			close(buckets)
			return nil
		}

		cursor := b.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			select {
			case <-cancel:
				break
			default:
				if v == nil {
					buckets <- &bucketView{string(k), append(a.path, string(k)), a.db}
				}
			}
		}
		close(buckets)

		return nil
	})

	return <-c
}

func (a *bucketView) BucketView(path ...string) BucketView {
	if len(path) == 0 {
		return a
	}

	b := make(chan BucketView, 1)
	a.db.View(func(tx *bolt.Tx) error {
		parent := bucket(tx, nil, a.path)
		if parent == nil {
			b <- nil
			return nil
		}
		target := bucket(tx, parent, path)
		if target == nil {
			b <- nil
			return nil
		}

		b <- &bucketView{path[len(path)-1], path, a.db}

		return nil
	})

	return <-b
}

func bucket(tx *bolt.Tx, parent *bolt.Bucket, path []string) *bolt.Bucket {
	if len(path) == 0 {
		return parent
	}

	if parent != nil {
		parent = parent.Bucket([]byte(path[0]))
	} else {
		parent = tx.Bucket([]byte(path[0]))
	}
	if parent == nil {
		return nil
	}
	return bucket(tx, parent, path[1:])
}
