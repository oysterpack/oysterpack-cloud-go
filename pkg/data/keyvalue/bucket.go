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

import "github.com/coreos/bbolt"

// Bucket represents a bucket of key-value pairs. Keys are strings, but values are simply bytes.
// Buckets can form a hierarchy of buckets.
type Bucket interface {
	BucketView

	// Put sets the value for a key in the bucket. If the key exist then its previous value will be overwritten.
	// Supplied value must remain valid for the life of the transaction.
	// Returns an error if the key is blank, if the key is too large, or if the value is too large.
	Put(key string, value []byte) error

	// PutMultiple will put all values received on the data channel within the same transaction, i.e., either all or none will be stored.
	// The puts are performed async, and when the the puts are done, then the result will be communicated via the response channel.
	// Once nil is received on the data channel, then this signals the transation was successfully committed.
	// If the transaction failed, then the error is returned on the response channel.
	PutMultiple(data <-chan *KeyValue) <-chan error

	// Delete removes the keys from the bucket. If the key does not exist then nothing is done and a nil error is returned.
	// All or none are deleted within the same transaction.
	Delete(keys ...string) error

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

type bucket struct {
	*bucketView
}

func (a *bucket) Put(key string, value []byte) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		b := lookupBucket(tx, nil, a.path)
		if b == nil {
			return errBucketDoesNotExist(a.path)
		}
		b.Put([]byte(key), value)
		return nil
	})
}

func (a *bucket) PutMultiple(data <-chan *KeyValue) <-chan error {
	c := make(chan chan error)

	go func() {
		response := make(chan error)
		defer close(response)
		c <- response

		err := a.db.Update(func(tx *bolt.Tx) error {
			b := lookupBucket(tx, nil, a.path)
			if b == nil {
				return errBucketDoesNotExist(a.path)
			}
			for kv := range data {
				b.Put([]byte(kv.Key), kv.Value)
			}

			return nil
		})

		if err != nil {
			response <- err
		}
	}()

	return <-c
}

func (a *bucket) Delete(keys ...string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		b := lookupBucket(tx, nil, a.path)
		if b == nil {
			return errBucketDoesNotExist(a.path)
		}
		for _, k := range keys {
			if err := b.Delete([]byte(k)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *bucket) CreateBucket(name string) (Bucket, error) {
	err := a.db.Update(func(tx *bolt.Tx) error {
		b := lookupBucket(tx, nil, a.path)
		if b == nil {
			return errBucketDoesNotExist(a.path)
		}
		_, err := b.CreateBucket([]byte(name))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &bucket{&bucketView{name: name, path: append(a.path, name), db: a.db}}, nil
}

func (a *bucket) CreateBucketIfNotExists(name string) (Bucket, error) {
	err := a.db.Update(func(tx *bolt.Tx) error {
		b := lookupBucket(tx, nil, a.path)
		if b == nil {
			return errBucketDoesNotExist(a.path)
		}
		_, err := b.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &bucket{&bucketView{name: name, path: append(a.path, name), db: a.db}}, nil
}

func (a *bucket) DeleteBucket(name string) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		b := lookupBucket(tx, nil, a.path)
		if b == nil {
			return errBucketDoesNotExist(a.path)
		}
		return b.DeleteBucket([]byte(name))
	})
}

func (a *bucket) Buckets(cancel <-chan struct{}) <-chan Bucket {
	c := make(chan chan Bucket)

	go a.db.View(func(tx *bolt.Tx) error {
		data := make(chan Bucket)
		c <- data
		b := lookupBucket(tx, nil, a.path)
		if b == nil {
			close(data)
			return nil
		}

		cursor := b.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			select {
			case <-cancel:
				break
			default:
				// nil values mean the value is a bucket
				if v == nil {
					data <- &bucket{&bucketView{string(k), append(a.path, string(k)), a.db}}
				}
			}
		}
		close(data)

		return nil
	})

	return <-c
}

func (a *bucket) Bucket(path ...string) Bucket {
	view := a.bucketView.bucketView(path...)
	if view != nil {
		return &bucket{view}
	}
	return nil
}
