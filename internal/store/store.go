package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"

	"github.com/yzoz/entropy/internal/domain"
)

var (
	bucketContracts = []byte("contracts")
	bucketAliases   = []byte("aliases")
	bucketPublics   = []byte("publics")
	bucketMeta      = []byte("meta")
	metaSchemaKey   = []byte("schema_version")
)

// schemaVersion is the on-disk layout version managed by Store.
const schemaVersion = 1

// Store is a bbolt-backed persistence layer for contracts and aliases.
type Store struct {
	db *bolt.DB
}

// Open opens (creating if necessary) the ledger database at path.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketContracts, bucketAliases, bucketPublics, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return tx.Bucket(bucketMeta).Put(metaSchemaKey, []byte(fmt.Sprint(schemaVersion)))
	})
}

// Close releases the database file.
func (s *Store) Close() error { return s.db.Close() }

// PutContract inserts or replaces a contract by its ID.
func (s *Store) PutContract(c *domain.Contract) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketContracts).Put([]byte(c.ID), b)
	})
}

// GetContract loads a contract by ID.
func (s *Store) GetContract(id string) (*domain.Contract, error) {
	var c domain.Contract
	err := s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketContracts).Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &c)
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListContracts returns every contract, ordered by ID.
func (s *Store) ListContracts() ([]*domain.Contract, error) {
	var out []*domain.Contract
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketContracts).ForEach(func(_, v []byte) error {
			var c domain.Contract
			if err := json.Unmarshal(v, &c); err != nil {
				return err
			}
			out = append(out, &c)
			return nil
		})
	})
	return out, err
}

// HasContract reports whether a contract with the given ID exists.
func (s *Store) HasContract(id string) (bool, error) {
	found := false
	err := s.db.View(func(tx *bolt.Tx) error {
		found = tx.Bucket(bucketContracts).Get([]byte(id)) != nil
		return nil
	})
	return found, err
}

// SetAlias stores a human-readable alias => address mapping.
func (s *Store) SetAlias(alias, address string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketAliases).Put([]byte(alias), []byte(address))
	})
}

// PutPublic records the hex public key for an address (needed to verify
// signatures on imported contracts).
func (s *Store) PutPublic(address, pubHex string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPublics).Put([]byte(address), []byte(pubHex))
	})
}

// GetPublic returns the hex public key for an address, if known.
func (s *Store) GetPublic(address string) (string, bool) {
	var pub string
	found := false
	_ = s.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket(bucketPublics).Get([]byte(address)); v != nil {
			pub = string(v)
			found = true
		}
		return nil
	})
	return pub, found
}

// ListPublics returns the address => public key map.
func (s *Store) ListPublics() (map[string]string, error) {
	out := map[string]string{}
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPublics).ForEach(func(k, v []byte) error {
			out[string(k)] = string(v)
			return nil
		})
	})
	return out, err
}

// ListAliases returns the alias => address map.
func (s *Store) ListAliases() (map[string]string, error) {
	out := map[string]string{}
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketAliases).ForEach(func(k, v []byte) error {
			out[string(k)] = string(v)
			return nil
		})
	})
	return out, err
}

// ResolveAlias maps an alias to an address, or returns the input unchanged if
// it is already an address.
func (s *Store) ResolveAlias(name string) string {
	addr := name
	_ = s.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket(bucketAliases).Get([]byte(name)); v != nil {
			addr = string(v)
		}
		return nil
	})
	return addr
}

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")
