package dumb

import (
	"sync"
)

type pubStore struct {
}

func NewPubStore(wg *sync.WaitGroup, shutdownCh <-chan struct{}, debug bool) *pubStore {
	return &pubStore{}
}

func (this *pubStore) Start() (err error) {
	return
}

func (this *pubStore) SyncPub(cluster string, topic, key string,
	msg []byte) (partition int32, offset int64, err error) {
	return
}

func (this *pubStore) AsyncPub(cluster string, topic, key string,
	msg []byte) (partition int32, offset int64, err error) {

	return
}