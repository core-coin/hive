package suite_cancun

import (
	"testing"

	"github.com/core-coin/go-core/common"
)

func TestBeaconRootStorageIndexes(t *testing.T) {

	expectedTimestampKey := common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000000a")
	expectedRootKey := common.HexToHash("0x000000000000000000000000000000000000000000000000000000000001800a")

	gotTimestampKey, gotRootKey := BeaconRootStorageIndexes(0xa)

	if gotTimestampKey != expectedTimestampKey {
		t.Fatal("expected timestamp key to be", expectedTimestampKey.Hex(), "got", gotTimestampKey.Hex())
	}
	if gotRootKey != expectedRootKey {
		t.Fatal("expected root key to be", expectedRootKey.Hex(), "got", gotRootKey.Hex())
	}
}
