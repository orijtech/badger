package badger

import (
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dgraph-io/badger/options"
)

func TestPersistentCache_DirectBadger(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	config := DefaultOptions
	config.TableLoadingMode = options.MemoryMap
	config.ValueLogFileSize = 16 << 20
	config.LevelOneSize = 8 << 20
	config.MaxTableSize = 2 << 20
	config.Dir = dir
	config.ValueDir = dir
	config.SyncWrites = false

	db, err := Open(config)
	if err != nil {
		t.Fatalf("cannot open db at location %s: %v", dir, err)
	}

	atomic.LoadUint64(&db.orc.curRead)

	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
}

type orc struct {
	isManaged bool // Does not change value, so no locking required.

	sync.Mutex
	curRead    uint64
	nextCommit uint64

	// These two structures are used to figure out when a commit is done. The minimum done commit is
	// used to update curRead.
	commitMark     uint64Heap
	pendingCommits map[uint64]struct{}

	// commits stores a key fingerprint and latest commit counter for it.
	// refCount is used to clear out commits map to avoid a memory blowup.
	commits  map[uint64]uint64
	refCount int64
}

func TestAtomicLoadArm(t *testing.T) {
	o := &orc{}
	atomic.LoadUint64(&o.curRead)
}
