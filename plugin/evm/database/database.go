// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package database

import (
	"fmt"
	"path/filepath"

	"github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leveldb"
	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ava-labs/avalanchego/database/meterdb"
	"github.com/ava-labs/avalanchego/database/pebbledb"
	"github.com/ava-labs/avalanchego/database/versiondb"
	avalanchenode "github.com/ava-labs/avalanchego/node"
	"github.com/ava-labs/avalanchego/utils/logging"
)

const (
	dbMetricsPrefix = "db"
)

// createDatabase returns a new database instance with the provided configuration
func NewStandaloneDatabase(dbConfig avalanchenode.DatabaseConfig, gatherer metrics.MultiGatherer, logger logging.Logger) (database.Database, error) {
	dbRegisterer, err := metrics.MakeAndRegister(
		gatherer,
		dbMetricsPrefix,
	)
	if err != nil {
		return nil, err
	}
	var db database.Database
	// start the db
	switch dbConfig.Name {
	case leveldb.Name:
		dbPath := filepath.Join(dbConfig.Path, leveldb.Name)
		db, err = leveldb.New(dbPath, dbConfig.Config, logger, dbRegisterer)
		if err != nil {
			return nil, fmt.Errorf("couldn't create %s at %s: %w", leveldb.Name, dbPath, err)
		}
	case memdb.Name:
		db = memdb.New()
	case pebbledb.Name:
		dbPath := filepath.Join(dbConfig.Path, pebbledb.Name)
		db, err = pebbledb.New(dbPath, dbConfig.Config, logger, dbRegisterer)
		if err != nil {
			return nil, fmt.Errorf("couldn't create %s at %s: %w", pebbledb.Name, dbPath, err)
		}
	default:
		return nil, fmt.Errorf(
			"db-type was %q but should have been one of {%s, %s, %s}",
			dbConfig.Name,
			leveldb.Name,
			memdb.Name,
			pebbledb.Name,
		)
	}

	if dbConfig.ReadOnly && dbConfig.Name != memdb.Name {
		db = versiondb.New(db)
	}

	meterDBReg, err := metrics.MakeAndRegister(
		gatherer,
		"meterdb",
	)
	if err != nil {
		return nil, err
	}

	db, err = meterdb.New(meterDBReg, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create meterdb: %w", err)
	}

	return db, nil
}
