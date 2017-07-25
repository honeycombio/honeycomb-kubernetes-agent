package tailer

import (
	"errors"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

const bucketName = "honeycomb-agent-state"

type StateRecorder interface {
	Record(path string, offset int64) error
	Get(path string) (int64, error)
	Delete(path string) error
}

type StateRecorderImpl struct {
	db *bolt.DB
}

func NewStateRecorder(stateFilePath string) (*StateRecorderImpl, error) {
	db, err := bolt.Open(stateFilePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	return &StateRecorderImpl{
		db: db,
	}, nil
}

func (s *StateRecorderImpl) Record(path string, offset int64) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			b, err := tx.CreateBucket([]byte(bucketName))
			if err != nil {
				return err
			}
			bucket = b
		}
		err := bucket.Put([]byte(path), []byte(strconv.FormatInt(offset, 10)))
		return err
	})
	return err
}

func (s *StateRecorderImpl) Get(path string) (offset int64, err error) {
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			err = errors.New("bucket not found")
			return err
		}
		v := b.Get([]byte(path))
		if v == nil {
			err = errors.New("key not found")
			return err
		}
		offset, err = strconv.ParseInt(string(v), 10, 0)
		return err
	})
	return offset, err
}

func (s *StateRecorderImpl) Delete(path string) (err error) {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(path))
	})
}
