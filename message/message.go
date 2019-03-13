package message

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"entropy/utils"

	"golang.org/x/crypto/ed25519"
)

//Message format
type Message struct {
	To   string
	From string
	Type string
	Amm  float64
	Curr string
	Exp  int64
	Sign string
}

func sign(p, m []byte) (s []byte) {
	s = ed25519.Sign(p, m)
	return
}

//Send ...
func Send(r, k, t, c string, amm float64, exp int64) {
	f := utils.User()
	private, err := ioutil.ReadFile(f + "/" + k + ".private")
	if err != nil {
		panic(err)
	}
	public, err := ioutil.ReadFile(f + "/" + k + ".public")
	if err != nil {
		panic(err)
	}

	e := strconv.FormatInt(exp, 10)
	//a, err := strconv.ParseFloat(amm, 64)
	a := strconv.FormatFloat(amm, 'f', -1, 64)
	priv := utils.Decode(string(private))
	p := string(public)

	msg := []byte(r + p + t + a + c + e)
	s := hex.EncodeToString(sign(priv, msg))

	m := Message{r, p, t, amm, c, exp, s}
	j, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))
}
