package tests

import (
	"bytes"
	"encoding/hex"
	"os"
	"pvsimflowtracking"
	"testing"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

// Example Test
func TestBase(t *testing.T) {
	t.Logf("This is an example test: %s", t.Name())
	ft := pvsimflowtracking.FlowTracking{}

	t.Logf("ft struct: %v: ", ft)

}

// Example Fuzzing Test
func FuzzHex(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {9}, {0xa}, {0xf}, {1, 2, 3, 4}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		enc := hex.EncodeToString(in)
		out, err := hex.DecodeString(enc)
		if err != nil {
			t.Fatalf("%v: decode: %v", in, err)
		}
		if !bytes.Equal(in, out) {
			t.Fatalf("%v: not equal after round trip: %v", in, out)
		}
	})
}
