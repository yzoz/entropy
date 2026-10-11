package bundle

import (
	"strings"
	"testing"
	"time"

	"github.com/yzoz/entropy/internal/domain"
)

func TestEncodeCSV(t *testing.T) {
	contracts := []*domain.Contract{
		{
			Version: domain.ContractVersion, ID: "abc", Type: domain.IOU,
			From: "ed25519:alice", To: "ed25519:bob", Amount: 1250, Currency: domain.USD,
			IssuedAt: time.Unix(1700000000, 0).Unix(), ExpiresAt: time.Unix(1700086400, 0).Unix(),
			Status: domain.Signed, Signatures: []domain.Signature{{Signer: "ed25519:alice"}},
			Memo: "lunch",
		},
	}
	raw, err := EncodeCSV(contracts)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.HasPrefix(s, "id,type,from,to,amount,currency") {
		t.Fatalf("bad header: %q", s)
	}
	for _, want := range []string{"abc", "IOU", "ed25519:alice", "12.50", "USD", "signed"} {
		if !strings.Contains(s, want) {
			t.Errorf("CSV missing %q in %q", want, s)
		}
	}
	// A single contract plus header means two lines.
	if lines := strings.Count(s, "\n"); lines != 2 {
		t.Errorf("expected 2 lines, got %d", lines)
	}
}
