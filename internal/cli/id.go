package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/crypto"
)

func newIDCmd(opt *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "Manage participant aliases",
	}
	cmd.AddCommand(newIDSetCmd(opt), newIDListCmd(opt))
	return cmd
}

func newIDSetCmd(opt *options) *cobra.Command {
	var name, address, key string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Associate an alias with an address",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				addr := address
				if addr == "" && key != "" {
					pub, err := publicForLocalKey(opt.Home, key)
					if err != nil {
						return err
					}
					addr = crypto.Address(pub)
				}
				if addr == "" {
					return fmt.Errorf("provide --address or --key")
				}
				if _, err := crypto.ParseAddress(addr); err != nil {
					return err
				}
				if err := a.st.SetAlias(name, addr); err != nil {
					return err
				}
				if opt.JSON {
					a.printJSON(map[string]string{"alias": name, "address": addr})
					return nil
				}
				fmt.Printf("Alias %q -> %s\n", name, addr)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "alias name (required)")
	cmd.Flags().StringVar(&address, "address", "", "address to bind to the alias")
	cmd.Flags().StringVar(&key, "key", "", "local key name whose address to bind")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newIDListCmd(opt *options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List known aliases",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				aliases, err := a.st.ListAliases()
				if err != nil {
					return err
				}
				if opt.JSON {
					a.printJSON(aliases)
					return nil
				}
				if len(aliases) == 0 {
					fmt.Println("No aliases.")
					return nil
				}
				for name, addr := range aliases {
					fmt.Printf("%-16s %s\n", name, addr)
				}
				return nil
			})
		},
	}
}
