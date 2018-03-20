/*
 * Copyright 2017 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package badger

import (
	octrace "go.opencensus.io/trace"
	"golang.org/x/net/context"
)

// ManagedDB allows end users to manage the transactions themselves. Transaction
// start and commit timestamps are set by end-user.
//
// This is only useful for databases built on top of Badger (like Dgraph), and
// can be ignored by most users.
//
// WARNING: This is an experimental feature and may be changed significantly in
// a future release. So please proceed with caution.
type ManagedDB struct {
	*DB
}

// OpenManaged returns a new ManagedDB, which allows more control over setting
// transaction timestamps.
//
// This is only useful for databases built on top of Badger (like Dgraph), and
// can be ignored by most users.
func OpenManaged(ctx context.Context, opts Options) (*ManagedDB, error) {
	ctx, span := octrace.StartSpan(ctx, "OpenManaged")
	defer span.End()
	opts.managedTxns = true
	db, err := Open(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &ManagedDB{db}, nil
}

// NewTransaction overrides DB.NewTransaction() and panics when invoked. Use
// NewTransactionAt() instead.
func (db *ManagedDB) NewTransaction(update bool) {
	panic("Cannot use NewTransaction() for ManagedDB. Use NewTransactionAt() instead.")
}

// NewTransactionAt follows the same logic as DB.NewTransaction(), but uses the
// provided read timestamp.
//
// This is only useful for databases built on top of Badger (like Dgraph), and
// can be ignored by most users.
func (db *ManagedDB) NewTransactionAt(ctx context.Context, readTs uint64, update bool) *Txn {
	ctx, span := octrace.StartSpan(ctx, "ManagedDB.NewTransactionAt")
	txn := db.DB.NewTransaction(ctx, update)
	txn.readTs = readTs
	span.End()
	return txn
}

// CommitAt commits the transaction, following the same logic as Commit(), but
// at the given commit timestamp. This will panic if not used with ManagedDB.
//
// This is only useful for databases built on top of Badger (like Dgraph), and
// can be ignored by most users.
func (txn *Txn) CommitAt(ctx context.Context, commitTs uint64, callback func(error)) error {
	ctx, span := octrace.StartSpan(ctx, "Txn.CommitAt")
	defer span.End()

	if !txn.db.opt.managedTxns {
		return ErrManagedTxn
	}
	txn.commitTs = commitTs
	return txn.Commit(ctx, callback)
}

// PurgeVersionsBelow will delete all versions of a key below the specified version
func (db *ManagedDB) PurgeVersionsBelow(ctx context.Context, key []byte, ts uint64) error {
	txn := db.NewTransactionAt(ctx, ts, false)
	defer txn.Discard(ctx)
	return db.purgeVersionsBelow(ctx, txn, key, ts)
}

// GetSequence is not supported on ManagedDB. Calling this would result
// in a panic.
func (db *ManagedDB) GetSequence(_ []byte, _ uint64) (*Sequence, error) {
	panic("Cannot use GetSequence for ManagedDB.")
}
