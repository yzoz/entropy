package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"time"

	"golang.org/x/crypto/ed25519"
)

const layout = "2006-01-02T15:04:05"

//Timestamp ...
func Timestamp(d string) string {
	t, err := time.Parse(layout, d)
	if err != nil {
		panic(err)
	}
	return strconv.FormatInt(t.Local().Unix(), 10)
}

//DateTime ... d := utils.DateTime(time.Now().Unix()+86400)
func DateTime(d int64) string {
	t := time.Unix(d, 0).Format(layout)
	return t
}

//Encode ...
func Encode(k []byte) (e string) {
	e = hex.EncodeToString(k)
	return
}

//Decode ...
func Decode(k string) []byte {
	d, err := hex.DecodeString(k)
	if err != nil {
		panic(err)
	}
	return d
}

//User ...
func User() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	f := usr.HomeDir + "/.entropy"
	return f
}

//Gen ...
func Gen(k string) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	f := User()
	if err != nil {
		panic(err)
	}
	ePriv := Encode(priv)
	ePub := Encode(pub)
	fmt.Println("PRIVATE :", ePriv)
	fmt.Println("PUBLIC :", ePub)

	os.Mkdir(f, 0755)

	ioutil.WriteFile(f+"/"+k+".private", []byte(ePriv), 0755)
	ioutil.WriteFile(f+"/"+k+".public", []byte(ePub), 0755)
}
