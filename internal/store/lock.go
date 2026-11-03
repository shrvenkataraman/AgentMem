package store

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// lockName is the file that serializes ingest runs inside a state dir.
const lockName = "ingest.lock"

// lockPoll is how often TryLock retries a busy lock.
const lockPoll = 10 * time.Millisecond

// defaultLockWait is used when AGENTMEM_LOCK_WAIT_MS is unset or malformed.
const defaultLockWait = 2000 * time.Millisecond

// LockWait is how long callers should let TryLock wait: $AGENTMEM_LOCK_WAIT_MS milliseconds,
// default 2000.
func LockWait() time.Duration {
	ms, err := strconv.Atoi(os.Getenv("AGENTMEM_LOCK_WAIT_MS"))
	if err != nil || ms < 0 {
		return defaultLockWait
	}
	return time.Duration(ms) * time.Millisecond
}

// TryLock takes an exclusive flock on <dir>/ingest.lock (created 0600), retrying every 10ms until
// wait elapses. It returns the unlock function and true when the lock was taken, or a nil unlock
// and false when it stayed busy. The caller must call unlock when done.
func TryLock(dir string, wait time.Duration) (unlock func(), ok bool) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, false
	}
	f, err := os.OpenFile(filepath.Join(dir, lockName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false
	}
	deadline := time.Now().Add(wait)
	for {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
			return func() { f.Close() }, true
		}
		if !time.Now().Before(deadline) {
			f.Close()
			return nil, false
		}
		time.Sleep(lockPoll)
	}
}
