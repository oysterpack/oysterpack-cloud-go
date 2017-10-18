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

	"fmt"

	"github.com/nats-io/nuid"
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

	if _, err := keyvalue.CreateDatabase(" ", "test", true); err == nil {
		t.Errorf("Should have failed because the file path is blank")
	} else {
		t.Log(err)
	}
	if _, err := keyvalue.CreateDatabase(dbFile, " ", true); err == nil {
		t.Errorf("Should have failed because the file path is blank")
	} else {
		t.Log(err)
	}
}

func TestCreateDatabase2(t *testing.T) {
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
	db.Close()

	if db, err := keyvalue.CreateDatabase(dbFile, "test2", false); err != nil {
		t.Error(err)
	} else {
		db.Close()
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

const (
	ACCOUNTS = "accounts"

	CREATED = "created"
	NAME    = "name"
)

/*
Create the following bucket structure :

	[accounts]
	 |-[{account-id}]
	    |- created
        |- name

*/
func TestDatabase_ReadWriteOperations(t *testing.T) {
	db := createTestDatabase(t, "TestDatabase_ReadWriteOperations", "test")
	defer db.Close()

	// create the 'accounts' bucket
	testCreateTopLevelDatabaseBucket(t, db, ACCOUNTS)
	accounts := db.Bucket(ACCOUNTS)
	if len(accounts.Path()) != 2 || accounts.Path()[0] != "test" || accounts.Path()[1] != ACCOUNTS {
		t.Fatal("The accounts bucket path should be [test -> accounts] : %s", accounts)
	}

	testCreateBucket(t, db, accounts)
	testDeleteBucket(t, db, accounts)
	testLookupBucketThatDoesNotExist(t, db)
}

func TestBucket_Buckets(t *testing.T) {
	db := createTestDatabase(t, "TestBucket_Buckets", "test")
	defer db.Close()

	// create the 'accounts' bucket
	testCreateTopLevelDatabaseBucket(t, db, ACCOUNTS)
	accounts := db.Bucket(ACCOUNTS)
	buckets := []keyvalue.Bucket{}
	for i := 0; i < 10; i++ {
		bucket, err := accounts.CreateBucket(nuid.Next())
		if err != nil {
			t.Fatal(err)
		}
		buckets = append(buckets, bucket)
	}
	cancelChan := make(chan struct{})
	returnedBuckets := []keyvalue.Bucket{}
	for bucket := range accounts.Buckets(cancelChan) {
		returnedBuckets = append(returnedBuckets, bucket)
	}
	if len(returnedBuckets) != len(buckets) {
		t.Errorf("returned count does not match : %d != %d", len(returnedBuckets), len(buckets))
	}

	keys := map[string]bool{}
	for _, bucket := range buckets {
		keys[bucket.Name()] = true
	}
	for bucket := range accounts.Buckets(nil) {
		delete(keys, bucket.Name())
	}
	if len(keys) != 0 {
		t.Errorf("No all keys have been returned : %v", keys)
	}

	close(cancelChan)
	<-accounts.Buckets(cancelChan)

	db.DeleteBucket(accounts.Name())
	<-accounts.Buckets(cancelChan)
}

func testCreateBucket(t *testing.T, db keyvalue.Database, accounts keyvalue.Bucket) {
	if _, err := accounts.CreateBucket(""); err == nil {
		t.Error("Should have failed because blank name is not allowed")
	}
	if _, err := accounts.CreateBucket("   "); err == nil {
		t.Error("Should have failed because blank name is not allowed")
	}
	if _, err := accounts.CreateBucketIfNotExists("  "); err == nil {
		t.Error("Should have failed because blank name is not allowed")
	}

	accountId := nuid.Next()
	account, err := accounts.CreateBucket(accountId)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name() != accountId {
		t.Errorf("*** ERROR *** account id does not match : %s : %s", account.Name())
	}

	testPutGetFields(t, account)
	testBucketView_Keys(t, account)
	testBucketView_KeyValues(t, account)

	account = accounts.Bucket(account.Name())
	if account == nil {
		t.Fatal("*** ERROR *** account bucket was not found")
	}
	if len(account.Path()) != 3 || account.Path()[0] != "test" || account.Path()[1] != ACCOUNTS || account.Path()[2] != accountId {
		t.Fatal("account path is wrong [%s], but it should be [test -> accounts -> %s]", account, accountId)
	}

	account = db.Bucket(ACCOUNTS, accountId)
	if account == nil || !account.Exists() {
		t.Fatal("account bucket was not found")
	}
	if len(account.Path()) != 3 || account.Path()[0] != "test" || account.Path()[1] != ACCOUNTS || account.Path()[2] != accountId {
		t.Fatal("account path is wrong [%s], but it should be [test -> accounts -> %s]", account, accountId)
	}

	if bucketView := accounts.BucketView(account.Name()); bucketView == nil {
		t.Error("bucket view should have been found")
	} else {
		t.Logf("bucketView : %v", bucketView)
		if bucketView.Name() != account.Name() {
			t.Errorf("name does not match : %s", bucketView.Name())
		}
		if fmt.Sprint(bucketView.Path()) != fmt.Sprint(account.Path()) {
			t.Errorf("path does not match : %s", bucketView)
		}
	}

	if accountsView := accounts.BucketView(); accountsView == nil {
		t.Error("it should have returned itself")
	} else {
		if fmt.Sprint(accountsView.Path()) != fmt.Sprint(accounts.Path()) {
			t.Error("it should have returned itself")
		}
	}

	testBucketView_BucketViews(t, accounts, 1)

	account, err = accounts.CreateBucketIfNotExists(account.Name())
	if err != nil {
		t.Error("Should not have failed")
	} else {
		if account.Path()[0] != db.Name() && account.Path()[1] != accounts.Name() && account.Path()[2] != accountId {
			t.Error("path does not match : %v : %s -> %s", account.Path(), accounts.Name(), accountId)
		}
	}
}

func testDeleteBucket(t *testing.T, db keyvalue.Database, accounts keyvalue.Bucket) {
	if err := db.DeleteBucket("  "); err == nil {
		t.Error("This should have failed because it is invalid to delete a bucket with a blank name")
	}

	keys := []string{}
	for key := range accounts.Keys("", nil) {
		keys = append(keys, key)
	}
	for _, key := range keys {
		accounts.DeleteBucket(key)
	}

	accountId := nuid.Next()
	account, err := accounts.CreateBucket(accountId)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name() != accountId {
		t.Errorf("*** ERROR *** account id does not match : %s : %s", account.Name())
	}

	testPutGetFields(t, account)
	testBucketView_Keys(t, account)
	testBucketView_KeyValues(t, account)

	account = accounts.Bucket(account.Name())
	if account == nil {
		t.Fatal("*** ERROR *** account bucket was not found")
	}
	if len(account.Path()) != 3 || account.Path()[0] != "test" || account.Path()[1] != ACCOUNTS || account.Path()[2] != accountId {
		t.Fatal("account path is wrong [%s], but it should be [test -> accounts -> %s]", account, accountId)
	}

	account = db.Bucket(ACCOUNTS, accountId)
	if account == nil || !account.Exists() {
		t.Fatal("account bucket was not found")
	}
	if len(account.Path()) != 3 || account.Path()[0] != "test" || account.Path()[1] != ACCOUNTS || account.Path()[2] != accountId {
		t.Fatal("account path is wrong [%s], but it should be [test -> accounts -> %s]", account, accountId)
	}

	if bucketView := accounts.BucketView(account.Name()); bucketView == nil {
		t.Error("bucket view should have been found")
	} else {
		t.Logf("bucketView : %v", bucketView)
		if bucketView.Name() != account.Name() {
			t.Errorf("name does not match : %s", bucketView.Name())
		}
		if fmt.Sprint(bucketView.Path()) != fmt.Sprint(account.Path()) {
			t.Errorf("path does not match : %s", bucketView)
		}
	}

	if accountsView := accounts.BucketView(); accountsView == nil {
		t.Error("it should have returned itself")
	} else {
		if fmt.Sprint(accountsView.Path()) != fmt.Sprint(accounts.Path()) {
			t.Error("it should have returned itself")
		}
	}

	testBucketView_BucketViews(t, accounts, 1)

	testBucket_DeleteBucket(t, accounts, account)

	if bucketView := accounts.BucketView(account.Name()); bucketView != nil {
		t.Errorf("bucket view should not have been found : %v : exists = %v", bucketView, bucketView.Exists())
	}

	testBucketView_BucketViews(t, accounts, 0)

	account, _ = accounts.CreateBucket(nuid.Next())
	testPutMultiple(t, account)
	account.Put("info", []byte("awesome"))
	if err = account.Delete(NAME, CREATED, nuid.Next()); err != nil {
		t.Error(err)
	}
	testPutMultipleInBucketThatDoesNotExist(t, accounts)
	testDeletingKeysFromBucketThatDoesNotExist(t, accounts)

	// delete the bucket
	if err := db.DeleteBucket(accounts.Name()); err != nil {
		t.Error(err)
	}

	// try operations on the deleted bucket
	testBucketView_BucketViews(t, accounts, 0)
	if accounts.BucketView(account.Name()) != nil {
		t.Error("should have returned nil because accounts was deleted")
	}

	<-accounts.KeyValues("", nil)
	if _, err := accounts.CreateBucket(nuid.Next()); err == nil {
		t.Error("should have failed because the bucket has been deleted")
	}

	if _, err := accounts.CreateBucketIfNotExists(nuid.Next()); err == nil {
		t.Error("should have failed because the bucket has been deleted")
	}
	if err := accounts.DeleteBucket(nuid.Next()); err == nil {
		t.Error("should have failed because the bucket has been deleted")
	}

}

func testDeletingKeysFromBucketThatDoesNotExist(t *testing.T, accounts keyvalue.Bucket) {
	account, err := accounts.CreateBucket(nuid.Next())
	if err != nil {
		t.Fatal(err)
	}
	err = accounts.DeleteBucket(account.Name())
	if err != nil {
		t.Fatal(err)
	}
	if err = account.Delete(NAME, CREATED, nuid.Next()); err == nil {
		t.Error("Error should have been returned because account bucket was deleted")
	} else {
		t.Log(err)
	}
}

func testPutMultipleInBucketThatDoesNotExist(t *testing.T, accounts keyvalue.Bucket) {
	account, err := accounts.CreateBucket(nuid.Next())
	if err != nil {
		t.Fatal(err)
	}
	err = accounts.DeleteBucket(account.Name())
	if err != nil {
		t.Fatal(err)
	}
	putChan := make(chan *keyvalue.KeyValue, 2)
	createdTS := time.Now()
	created, _ := createdTS.MarshalBinary()
	putChan <- &keyvalue.KeyValue{CREATED, created}
	putChan <- &keyvalue.KeyValue{NAME, []byte("Apple")}
	close(putChan)
	err = <-account.PutMultiple(putChan)
	if err == nil {
		t.Error("Expected error because bucket does not exist")
	} else {
		t.Log(err)
	}

	if err = account.Put(NAME, []byte("Google")); err == nil {
		t.Error("Expected error because bucket does not exist")
	} else {
		t.Log(err)
	}
}

func testPutMultiple(t *testing.T, account keyvalue.Bucket) {
	putChan := make(chan *keyvalue.KeyValue, 2)
	createdTS := time.Now()
	created, _ := createdTS.MarshalBinary()
	putChan <- &keyvalue.KeyValue{CREATED, created}
	putChan <- &keyvalue.KeyValue{NAME, []byte("Apple")}
	close(putChan)
	err := <-account.PutMultiple(putChan)
	if err != nil {
		t.Error(err)
	}
	if account.Get(CREATED) == nil || account.Get(NAME) == nil {
		t.Errorf("Fields were not stored properly : created : %v name : %v", account.Get(CREATED), account.Get(NAME))
	}
}

func testBucketView_BucketViews(t *testing.T, bucket keyvalue.Bucket, expectedChildCount int) {
	count := 0
	for child := range bucket.BucketViews(nil) {
		count++
		t.Log(child)
	}
	if count != expectedChildCount {
		t.Errorf("count != expectedChildCount : %d != %d", count, expectedChildCount)
	}

	cancel := make(chan struct{})
	count = 0
	for child := range bucket.BucketViews(cancel) {
		count++
		t.Log(child)
	}
	if count != expectedChildCount {
		t.Errorf("count != expectedChildCount : %d != %d", count, expectedChildCount)
	}

	close(cancel)
	count = 0
	for child := range bucket.BucketViews(cancel) {
		count++
		t.Log(child)
	}
	if count != 0 {
		t.Errorf("none should be returned because the request was cancelled")
	}
}

func testPutGetFields(t *testing.T, account keyvalue.Bucket) {
	createdTS := time.Now()
	created, _ := createdTS.MarshalBinary()
	account.Put(CREATED, created)
	account.Put(NAME, []byte("Amazon"))

	accountCreatedTS := time.Now()
	if err := accountCreatedTS.UnmarshalBinary(account.Get(CREATED)); err != nil {
		t.Errorf("Failed to unmarshal created time L %v", err)
	} else if !createdTS.Equal(accountCreatedTS) {
		t.Errorf("created time does not match : %v != %v", createdTS, accountCreatedTS)
	}

	if account.Get("") != nil {
		t.Error("No key should be blank")
	}
	if account.Get(nuid.Next()) != nil {
		t.Error("No key should be blank")
	}
}

func testBucket_DeleteBucket(t *testing.T, accounts keyvalue.Bucket, account keyvalue.Bucket) {
	if err := accounts.DeleteBucket(account.Name()); err != nil {
		t.Errorf("*** ERROR *** failed to delete bucket : %v", err)
	} else {
		if account.Exists() {
			t.Error("account is still reporting that it exists")
		}
		if account.Get("created") != nil {
			t.Error("Account was deleted - so it should have no data")
		}
		account.Keys("", nil)
	}
}

func testBucketView_Keys(t *testing.T, account keyvalue.Bucket) {
	keysChan := account.Keys("", make(chan struct{}))
	keys := []string{}
	for key := range keysChan {
		keys = append(keys, key)
	}
	if len(keys) != 2 || keys[0] != CREATED || keys[1] != NAME {
		t.Errorf("*** ERROR *** wrong number of keys")
	}
	keysChan = account.Keys("", nil)
	keys = []string{}
	for key := range keysChan {
		keys = append(keys, key)
	}
	if len(keys) != 2 || keys[0] != CREATED || keys[1] != NAME {
		t.Errorf("*** ERROR *** wrong number of keys")
	}

	keysChan = account.Keys("n", nil)
	keys = []string{}
	for key := range keysChan {
		keys = append(keys, key)
	}
	if len(keys) != 1 || keys[0] != NAME {
		t.Errorf("*** ERROR *** wrong number of keys")
	}

	cancelChan := make(chan struct{})
	close(cancelChan)
	keysChan = account.Keys("", cancelChan)
	keys = []string{}
	for key := range keysChan {
		keys = append(keys, key)
	}
	t.Logf("keys that were returned when the cancel channel was closed immediately : %v", keys)
	if len(keys) == 2 {
		t.Error("*** ERROR *** The cancel channel was closed immediately, so we don't expect all keys to be delivered")
	}
}

func testBucketView_KeyValues(t *testing.T, account keyvalue.Bucket) {
	kvChan := account.KeyValues("", make(chan struct{}))
	kvs := []*keyvalue.KeyValue{}
	for key := range kvChan {
		kvs = append(kvs, key)
	}
	if len(kvs) != 2 || kvs[0].Key != CREATED || kvs[1].Key != NAME || string(kvs[1].Value) != "Amazon" {
		t.Errorf("*** ERROR *** wrong number of keys %v", kvs)
	}
	kvChan = account.KeyValues("", nil)
	kvs = []*keyvalue.KeyValue{}
	for key := range kvChan {
		kvs = append(kvs, key)
	}
	if len(kvs) != 2 || kvs[0].Key != CREATED || kvs[1].Key != NAME {
		t.Errorf("*** ERROR *** wrong number of keys: %v", kvs)
	}

	kvChan = account.KeyValues("n", nil)
	kvs = []*keyvalue.KeyValue{}
	for key := range kvChan {
		kvs = append(kvs, key)
	}
	if len(kvs) != 1 || kvs[0].Key != NAME {
		t.Errorf("*** ERROR *** wrong number of keys")
	}

	cancelChan := make(chan struct{})
	close(cancelChan)
	kvChan = account.KeyValues("n", cancelChan)
	kvs = []*keyvalue.KeyValue{}
	for key := range kvChan {
		kvs = append(kvs, key)
	}
	t.Logf("key value pairs that were returned when the cancel channel was closed immediately : %v", kvs)
	if len(kvs) == 2 {
		t.Error("*** ERROR *** The cancel channel was closed immediately, so we don't expect all keys to be delivered")
	}
}

func testCreateTopLevelDatabaseBucket(t *testing.T, db keyvalue.Database, bucketName string) {
	accounts, err := db.CreateBucket(bucketName)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(accounts)
	if accounts.Name() != bucketName {
		t.Errorf("*** ERROR *** accounts bucket name does not match : %s", accounts.Name())
	}
	accounts = db.Bucket(accounts.Name())
	if accounts == nil {
		t.Fatal("*** ERROR *** accounts bucket was not found")
	}
	if len(accounts.Path()) != 2 || accounts.Path()[0] != "test" || accounts.Path()[1] != bucketName {
		t.Fatal("The accounts bucket path should be [test -> accounts] : %s", accounts)
	}

	if !accounts.Exists() {
		t.Fatal("accounts should exist")
	}
}

func testLookupBucketThatDoesNotExist(t *testing.T, db keyvalue.Database) {
	if db.Bucket(ACCOUNTS, nuid.Next()) != nil {
		t.Errorf("No bucket should have been found : %s", db.Bucket(ACCOUNTS, nuid.Next()))
	}
	if db.Bucket(ACCOUNTS, "") != nil {
		t.Errorf("No bucket should have been found : %s", db.Bucket(ACCOUNTS, nuid.Next()))
	}
}

func createTestDatabase(t *testing.T, test string, dbName string) keyvalue.Database {
	dbFile := dbFile(test)
	deleteDbFile(t, dbFile)
	if err := os.Remove(dbFile); err != nil {
		t.Log(err)
	}
	db, err := keyvalue.CreateDatabase(dbFile, dbName, true)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
