package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"
)

func main() {
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	fmt.Println(pubKey)
	fmt.Println(privKey)
	hexPub := hex.EncodeToString(pubKey)
	hexPriv := hex.EncodeToString(privKey)
	fmt.Println(hexPub)
	fmt.Println(hexPriv)
	strHexPub, _ := hex.DecodeString(hexPub)
	strHexPriv, _ := hex.DecodeString(hexPriv)
	fmt.Println(strHexPub)
	fmt.Println(strHexPriv)
	b2b, _ := blake2b.New(64, nil)
	m := []byte("qweqwe qeqweqw e t r ete ry e ty rt u ut  y ity i  i yu yu o  yo i o i p p  uip  uip")
	b2b.Write(m)
	fmt.Println(hex.EncodeToString(b2b.Sum(nil)))
	fmt.Println(base64.StdEncoding.EncodeToString(b2b.Sum(nil)))
	s := ed25519.Sign(privKey, m)
	ss := hex.EncodeToString(s)
	fmt.Println(ss)
	sb, _ := hex.DecodeString(ss)
	v := ed25519.Verify(pubKey, m, sb)
	fmt.Println(v)
}
