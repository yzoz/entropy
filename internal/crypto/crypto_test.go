package crypto

import (
	"testing"
)

func TestAddressRoundTrip(t *testing.T) {
	pub, _, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := Address(pub)
	if _, err := ParseAddress(addr); err != nil {
		t.Fatalf("derived address should parse: %v", err)
	}
	if _, err := ParseAddress("ed25519:not-base58!!"); err == nil {
		t.Error("garbage address should be rejected")
	}
}

func TestSignVerify(t *testing.T) {
	pub, priv, _ := GenerateKey()
	s := NewSigner(priv)
	msg := []byte("hello entropy")
	sig, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !(Ed25519Verifier{}).Verify(pub, msg, sig) {
		t.Error("valid signature did not verify")
	}
	if (Ed25519Verifier{}).Verify(pub, []byte("tampered"), sig) {
		t.Error("signature verified for the wrong message")
	}
}

func TestEncryptDecryptKey(t *testing.T) {
	pub, priv, _ := GenerateKey()
	privHex := EncodePrivate(priv)
	pubHex := EncodePublic(pub)

	blob, err := EncryptPrivateKey(privHex, "hunter2", pubHex)
	if err != nil {
		t.Fatal(err)
	}
	gotPriv, gotPub, err := DecryptPrivateKey(blob, "hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if gotPriv != privHex || gotPub != pubHex {
		t.Error("decrypted key does not match original")
	}
	if _, _, err := DecryptPrivateKey(blob, "wrong"); err == nil {
		t.Error("wrong passphrase should fail")
	}
}
