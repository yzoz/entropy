package main

import (
	"fmt"
	"strings"
	"time"

	"entropy/message"
	"entropy/utils"

	"github.com/spf13/cobra"
)

func main() {

	var Key string
	var Receiver string
	var Ammount float64
	var Currency string
	var Days int64
	var Hours int64
	var Minutes int64
	var Contract string
	var Type string
	var Direction string
	var Status string
	var From string

	var cmdGen = &cobra.Command{
		Use:   "gen",
		Short: "Generate keys",
		Long:  `Generate private and public keys to ~/.entropy folder`,
		Args:  cobra.NoArgs,
		Run: func(c *cobra.Command, a []string) {
			utils.Gen(Key)
		},
	}
	cmdGen.Flags().StringVarP(&Key, "key", "k", "key", "Name of the keys")

	var cmdSend = &cobra.Command{
		Use:   "send [contract type]",
		Short: "Send contract",
		Long:  `Send your contract to Chain {A | B | C} with default expration in 24 hours`,
		Args:  cobra.RangeArgs(1, 1),
		Run: func(c *cobra.Command, a []string) {
			var Exp int64
			Exp = Days*86400 + Hours*3600 + Minutes*60
			if Exp == 0 {
				Exp = 86400
			}
			Exp += time.Now().Unix()
			message.Send(Receiver, Key, strings.Join(a, " "), Currency, Ammount, Exp)
		},
	}
	cmdSend.Flags().StringVarP(&Receiver, "receiver", "r", "",
		"Receiver public key (required)")
	cmdSend.Flags().StringVarP(&Key, "key", "k", "key",
		"Your keys name")
	cmdSend.Flags().Float64VarP(&Ammount, "ammount", "a", 1,
		"Ammount to send (required)")
	cmdSend.Flags().StringVarP(&Currency, "currency", "c", "",
		"Currency of transactiom (required) {USD | EUR | RUR | GLD | SLV | OIL}")
	cmdSend.Flags().Int64VarP(&Days, "days", "d", 0,
		"Expiration in days")
	cmdSend.Flags().Int64VarP(&Hours, "hours", "o", 0,
		"Expiration in hours")
	cmdSend.Flags().Int64VarP(&Minutes, "minutes", "m", 0,
		"Expiration in minutes")
	cmdSend.MarkFlagRequired("receiver")
	cmdSend.MarkFlagRequired("ammount")
	cmdSend.MarkFlagRequired("currency")

	var cmdView = &cobra.Command{
		Use:   "view",
		Short: "View contracts",
		Long:  `View your contracts`,
		Args:  cobra.NoArgs,
		Run: func(c *cobra.Command, a []string) {
			fmt.Println("Contracts: " + Key)
		},
	}
	cmdView.Flags().StringVarP(&Key, "key", "k", "key", "Your public key name")
	cmdView.Flags().StringVarP(&Type, "type", "t", "", "Contract type {A | B | C}")
	cmdView.Flags().StringVarP(&Direction, "direction", "d", "", "sender | receiver")
	cmdView.Flags().StringVarP(&Status, "status", "s", "", "open | close")
	cmdView.Flags().StringVarP(&Currency, "currency", "c", "",
		"Currency of transactiom {USD | EUR | RUR | GLD | SLV | OIL}")
	cmdView.Flags().StringVarP(&From, "from", "f", utils.DateTime(time.Now().Unix()-86400), "From date in UTC {YYYY-MM-DDTHH:MM:SS}")

	var cmdSign = &cobra.Command{
		Use:   "sign",
		Short: "Sign contract",
		Long:  `Sign received contract`,
		Args:  cobra.NoArgs,
		Run: func(c *cobra.Command, a []string) {
			fmt.Println("Contract: ")
		},
	}
	cmdSign.Flags().StringVarP(&Contract, "id", "i", "", "Contract Chain ID")
	cmdSign.Flags().StringVarP(&Key, "key", "k", "key", "Your keys name")
	cmdSign.MarkFlagRequired("id")

	var rootCmd = &cobra.Command{
		Use:   "entropy",
		Short: "Entropy",
		Long:  `Entropy Money`,
	}
	rootCmd.AddCommand(cmdGen, cmdSend, cmdView, cmdSign)
	rootCmd.Execute()
}
