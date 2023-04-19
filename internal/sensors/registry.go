// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
)

type sensorRegistry struct {
	uri fyne.URI
	db  *badger.DB
}

func OpenSensorRegistry(ctx context.Context, appPath fyne.URI) *sensorRegistry {
	uri, err := storage.Child(appPath, "sensorRegistry")
	if err != nil {
		log.Error().Err(err).
			Msg("Unable open sensor registry path. Will not be able to track sensor status.")
		return nil
	}

	// Open a badgerDB with largely the default options, but trying to optimise
	// for lower memory usage as per:
	// https://dgraph.io/docs/badger/get-started/#memory-usage
	db, err := badger.Open(badger.DefaultOptions(uri.Path()).
		// * If the number of sensors is large, this might need adjustment.
		WithMemTableSize(12 << 20))
	if err != nil {
		log.Debug().Err(err).Msg("Could not open sensor registry DB.")
		return nil
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
		again:
			err := db.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	}()

	go func() {
		<-ctx.Done()
		log.Debug().Caller().Msg("Closing registry.")
		db.Close()
	}()

	return &sensorRegistry{
		uri: uri,
		db:  db,
	}
}

func (reg *sensorRegistry) CloseSensorRegistry() {
	reg.db.Close()
}

func (reg *sensorRegistry) Add(id string) *registryEntry {
	entry := newRegistryEntry(id)
	if values, err := reg.Get(entry.id); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			log.Debug().
				Msgf("Adding %s to registry DB.", entry.id)
			err := reg.Set(entry.id, entry.values)
			if err != nil {
				log.Debug().Err(err).
					Msgf("Could not add %s to registry DB.", id)
			}
		} else {
			log.Debug().Err(err).Msg("Could not retrieve state.")
		}
	} else {
		entry.values = values
	}
	return entry
}

func (reg *sensorRegistry) Get(id string) (*registryValues, error) {
	state := &registryValues{}
	err := reg.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, state)
			return err
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return state, err
	}
	return state, nil
}

func (reg *sensorRegistry) Set(id string, values *registryValues) error {
	err := reg.db.Update(func(txn *badger.Txn) error {
		v, err := json.Marshal(values)
		if err != nil {
			return err
		}
		err = reg.db.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(id), v)
			return err
		})
		return err
	})
	return err
}

func (reg *sensorRegistry) Update(entry *registryEntry) error {
	return reg.Set(entry.id, entry.values)
}

type registryValues struct {
	Registered bool `json:"Registered"`
	Disabled   bool `json:"Disabled"`
}

func newRegistryValues() *registryValues {
	return &registryValues{
		Disabled:   false,
		Registered: false,
	}
}

type registryEntry struct {
	id     string
	values *registryValues
}

func newRegistryEntry(id string) *registryEntry {
	return &registryEntry{
		id:     id,
		values: newRegistryValues(),
	}
}

func (e *registryEntry) IsDisabled() bool {
	return e.values.Disabled
}

func (e *registryEntry) SetDisabled(state bool) {
	e.values.Disabled = state
}

func (e *registryEntry) IsRegistered() bool {
	return e.values.Registered
}

func (e *registryEntry) SetRegistered(state bool) {
	e.values.Registered = state
}
