package radioserver

import (
	"github.com/luigifreitas/radioserver/protocol"
)

var ServerVersion = protocol.Version{
	Major: 0,
	Minor: 1,
	Hash:  0,
}

func init() {
	ServerVersion.Hash = uint32(122)
}
