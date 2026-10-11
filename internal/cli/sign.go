package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/domain"
)

func newSignCmd(opt *options) *cobra.Command {
	var id, keyName, passphrase string
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Countersign a received contract",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				c, err := a.st.GetContract(id)
				if err != nil {
					return fmt.Errorf("contract %q not found", id)
				}
				who, err := a.signer(keyName, passphrase)
				if err != nil {
					return err
				}
				if err := a.signService().Sign(c, who); err != nil {
					return err
				}
				if err := a.st.PutContract(c); err != nil {
					return err
				}
				if opt.JSON {
					a.printJSON(a.contractSummary(c))
					return nil
				}
				fmt.Printf("Signed %s -> %s\n", c.ID, c.Status)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "contract ID (required)")
	cmd.Flags().StringVarP(&keyName, "key", "k", "key", "local key name to sign with")
	cmd.Flags().StringVar(&passphrase, "passphrase", "", "key passphrase (omit to be prompted)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newSettleCmd(opt *options) *cobra.Command {
	var c = struct {
		key, passphrase, amount, currency, settles, memo string
	}{}
	cmd := &cobra.Command{
		Use:   "settle",
		Short: "Settle (repay) an outstanding IOU",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				debtor, err := a.signer(c.key, c.passphrase)
				if err != nil {
					return err
				}
				ref, err := a.st.GetContract(c.settles)
				if err != nil {
					return fmt.Errorf("referenced IOU %q not found", c.settles)
				}
				if ref.Type != domain.IOU {
					return fmt.Errorf("%s is not an IOU", c.settles)
				}
				if ref.From != debtor.Address {
					return fmt.Errorf("only the debtor (%s) can settle %s", ref.From, ref.ID)
				}
				amount := c.amount
				if amount == "" {
					amount = domain.FormatAmount(ref.Currency, ref.Amount)
				}
				currency := c.currency
				if currency == "" {
					currency = string(ref.Currency)
				}
				out, err := a.createContract(createParams{
					Type:       domain.SETTLE,
					FromKey:    c.key,
					Passphrase: c.passphrase,
					To:         ref.To,
					Amount:     amount,
					Currency:   currency,
					Memo:       c.memo,
					Settles:    ref.ID,
				})
				if err != nil {
					return err
				}
				if opt.JSON {
					a.printJSON(a.contractSummary(out))
					return nil
				}
				fmt.Println(a.shortLine(out) + "  (draft — awaiting creditor signature)")
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&c.settles, "iou", "i", "", "ID of the IOU being repaid (required)")
	cmd.Flags().StringVarP(&c.key, "key", "k", "key", "local key name to sign with")
	cmd.Flags().StringVar(&c.passphrase, "passphrase", "", "key passphrase (omit to be prompted)")
	cmd.Flags().StringVarP(&c.amount, "amount", "a", "", "amount to settle (defaults to the full IOU)")
	cmd.Flags().StringVarP(&c.currency, "currency", "c", "", "currency (defaults to the IOU currency)")
	cmd.Flags().StringVar(&c.memo, "memo", "", "optional note")
	cmd.MarkFlagRequired("iou")
	return cmd
}
