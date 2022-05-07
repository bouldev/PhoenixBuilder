package utils

import (
	"os"

	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/util"
)

func MakeDirP(path string) error {
	stat, err := os.Stat(path)
	if !(err == nil && stat.IsDir()) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if !(err == nil && stat.IsDir()) {
		return false
	}
	return true
}

func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if !(err == nil && !stat.IsDir()) {
		return false
	}
	return true
}

type LevelDBWrapper struct {
	*leveldb.DB
}

func (ldw *LevelDBWrapper) Get(key string) string {
	value, err := ldw.DB.Get([]byte(key), nil)
	if err != nil {
		return ""
	}
	return string(value)
}

func (ldw *LevelDBWrapper) Delete(key string) {
	ldw.DB.Delete([]byte(key), nil)
}

func (ldw *LevelDBWrapper) Commit(key string, v string) {
	err := ldw.Put([]byte(key), []byte(v), nil)
	if err != nil {
		panic(err)
	}
}

func (ldw *LevelDBWrapper) iter(cb func(key string, v string) bool, slice *util.Range) {
	iter := ldw.DB.NewIterator(slice, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		if cb(string(key), string(value)) {
			break
		}
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		panic(err)
	}
}

func (ldw *LevelDBWrapper) IterAll(cb func(key string, v string) bool) {
	ldw.iter(cb, nil)
}

func (ldw *LevelDBWrapper) IterWithPrefix(cb func(key string, v string) bool, prefix string) {
	ldw.iter(cb, util.BytesPrefix([]byte(prefix)))
}

func (ldw *LevelDBWrapper) IterWithRange(cb func(key string, v string) bool, start, end string) {
	ldw.iter(cb, &util.Range{
		Start: []byte(start),
		Limit: []byte(end),
	})
}

func GetLevelDB(path string) *LevelDBWrapper {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		panic(err)
	}
	return &LevelDBWrapper{DB: db}
}
