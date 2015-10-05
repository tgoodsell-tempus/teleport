package services

import (
	"github.com/gravitational/teleport/lib/backend/encryptedbk"
)

type BkKeysService struct {
	*encryptedbk.ReplicatedBackend
}

func NewBkKeysService(backend *encryptedbk.ReplicatedBackend) *BkKeysService {
	return &BkKeysService{backend}
}
