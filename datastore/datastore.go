package datastore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

var once sync.Once
var instance *driver

type driver struct {
	conn *badger.DB
}

type Datastore[T any] interface {
	// ReadAll - Read all the data in the DB, filter it by environment and
	// data type.
	// Use dataType "all" to return all data in the environment
	// Returns and array of UUIDs, or an error
	ReadAll(env string, dataType string) ([]string, error)

	// Read - Get an Object using its UUID, returns Object[T], or an error
	Read(uuid string) (Object[T], error)

	// Write - Create or update data, returns an error if any
	Write(uuid string, obj Object[T]) error

	// Delete - Delete an object from the database
	Delete(uuid string) error
}

type Object[T any] struct {
	// UUID: UUID v4 generated on creation
	// Rev: Revision, comprised of sequential int and a SHA256 hash of content
	// History: List of Rev IDs
	// Data: The data being stored
	// Type: Data type, used to find data by type
	// Environment: Used to silo data

	UUID        string   `json:"uuid"`
	Rev         string   `json:"_rev"`
	History     []string `json:"_history"`
	Data        T        `json:"data"`
	Environment string   `json:"_env"`
	Type        string   `json:"_type"`
}

type database[T any] struct {
	location string
}

func NewObject[T any](data T) (Object[T], error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return Object[T]{}, err
	}

	o := Object[T]{
		UUID:    uuid.String(),
		Rev:     "0-0", // Rev updated on write
		History: []string{},
		Data:    data,
	}

	return o, nil
}

func NewDatastore[T any](path string) Datastore[T] {
	db := &database[T]{
		location: path,
	}

	return db
}

func (d *database[T]) ReadAll(env string, dType string) ([]string, error) {
	var uuids = []string{}

	db, err := d.openConn(true)
	if err != nil {
		return uuids, err
	}

	err = db.conn.View(
		func(tx *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false

			it := tx.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()

				var b []byte
				if err := item.Value(
					func(val []byte) error {
						b = append(b, val...)
						return nil
					},
				); err != nil {
					return err
				}

				obj, err := convertToObject[T](b)
				if err != nil {
					return err
				}

				if obj.Environment != env {
					continue
				}

				if dType != "all" && obj.Type != dType {
					continue
				}

				uuids = append(uuids, obj.UUID)
			}

			return nil
		},
	)

	if err != nil {
		return uuids, err
	}

	return uuids, nil

}

func (d *database[T]) Read(uuid string) (Object[T], error) {
	var obj Object[T]

	db, err := d.openConn(true)
	if err != nil {
		return obj, err
	}

	vCopy := []byte{}

	if err := db.conn.View(
		func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(uuid))
			if err != nil {
				return err
			}

			if err := item.Value(
				// Copy value out
				func(val []byte) error {

					vCopy = append(vCopy, val...)

					return nil
				},
			); err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		return obj, err
	}

	obj, err = convertToObject[T](vCopy)
	if err != nil {
		return obj, err
	}

	return obj, nil
}

func (d *database[T]) Write(uuid string, obj Object[T]) error {

	rev, err := getRev(obj.Rev, obj)
	if err != nil {
		return err
	}

	obj.History = append(obj.History, rev)
	obj.Rev = rev

	bData, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	db, err := d.openConn(false)
	if err != nil {
		return err
	}

	if err := db.conn.Update(
		func(txn *badger.Txn) error {
			// create a byte buffer called bs
			bs := new(bytes.Buffer)
			_ = json.Indent(bs, bData, "", "  ")
			fmt.Println("Storing:")
			fmt.Println(bs.String())
			fmt.Println("UUID: ", uuid)

			return txn.Set([]byte(uuid), bData)
		},
	); err != nil {
		return err
	}

	return nil
}

func (d *database[T]) Delete(uuid string) error {

	db, err := d.openConn(false)
	if err != nil {
		return err
	}

	if err := db.conn.Update(
		func(txn *badger.Txn) error {
			return txn.Delete([]byte(uuid))
		},
	); err != nil {
		return err
	}

	return nil
}

func (d *database[T]) openConn(readonly bool) (*driver, error) {
	var connErr error

	once.Do(
		func() {
			opts := badger.DefaultOptions(d.location)
			opts.Logger = nil

			if readonly {
				opts.WithReadOnly(true)
			}

			conn, err := badger.Open(opts)
			if err != nil {
				connErr = err
				return
			}

			instance = &driver{
				conn,
			}
		},
	)

	if connErr != nil {
		return instance, connErr
	}

	return instance, nil
}

func convertToObject[T any](b []byte) (Object[T], error) {
	obj := Object[T]{}

	if err := json.Unmarshal(b, &obj); err != nil {
		return obj, err
	}

	return obj, nil
}

func getRev[T any](currRev string, o Object[T]) (string, error) {

	hasher := sha256.New()
	bData, err := json.Marshal(o.Data)
	if err != nil {
		return "", err
	}

	hasher.Write(bData)
	hash := hex.EncodeToString(hasher.Sum(nil))

	it := 1
	// Format <seq iter>-<content hash>
	// e.g: 2-50d858e0985ecc7f60418aaf0cc5ab587f42c2570a884095a9e8ccacd0f6545c
	if currRev != "" {
		currIt, err := strconv.ParseInt(strings.Split(currRev, "-")[0], 10, 0)
		if err != nil {
			return "", err
		}

		it = int(currIt) + 1
	}

	rev := fmt.Sprintf("%d-%s", it, hash)

	return rev, nil

}
