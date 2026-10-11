package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/domain"
)

func newIssueCmd(opt *options) *cobra.Command {
	return newCreateCmd(opt, domain.IOU, "issue", "Issue a debt note (IOU) in favour of a counterparty")
}

func newSendCmd(opt *options) *cobra.Command {
	return newCreateCmd(opt, domain.TRANSFER, "send", "Transfer credit to a counterparty (TRANSFER)")
}

// durationFlags holds expiration components shared by create-style commands.
type durationFlags struct {
	days, hours, minutes int64
}

func (d durationFlags) duration() time.Duration {
	return time.Duration(d.days)*24*time.Hour +
		time.Duration(d.hours)*time.Hour +
		time.Duration(d.minutes)*time.Minute
}

func newCreateCmd(opt *options, t domain.Type, use, short string) *cobra.Command {
	var p createParams
	var dur durationFlags
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p.Type = t
			p.Expires = dur.duration()
			a := &app{opt: opt}
			return a.withStore(func() error {
				c, err := a.createContract(p)
				if err != nil {
					return err
				}
				if opt.JSON {
					a.printJSON(a.contractSummary(c))
					return nil
				}
				note := ""
				if c.Status == domain.Draft {
					note = "  (draft — waiting for the counterparty to sign)"
				}
				fmt.Println(a.shortLine(c) + note)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&p.To, "to", "t", "", "counterparty alias or address (required)")
	cmd.Flags().StringVarP(&p.Amount, "amount", "a", "", "amount, e.g. 12.50 (required)")
	cmd.Flags().StringVarP(&p.Currency, "currency", "c", "USD", "currency {USD|EUR|RUR|GLD|SLV|OIL}")
	cmd.Flags().StringVarP(&p.FromKey, "key", "k", "key", "local key name to sign with")
	cmd.Flags().StringVar(&p.Passphrase, "passphrase", "", "key passphrase (omit to be prompted)")
	cmd.Flags().StringVar(&p.Memo, "memo", "", "optional note")
	cmd.Flags().Int64VarP(&dur.days, "days", "d", 0, "expiration in days")
	cmd.Flags().Int64VarP(&dur.hours, "hours", "o", 0, "expiration in hours")
	cmd.Flags().Int64VarP(&dur.minutes, "minutes", "m", 0, "expiration in minutes")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("amount")
	return cmd
}
