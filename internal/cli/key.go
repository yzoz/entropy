package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/yzoz/entropy/internal/crypto"
)

func newKeyCmd(opt *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage signing keys",
	}
	cmd.AddCommand(newKeyGenCmd(opt))
	return cmd
}

func newKeyGenCmd(opt *options) *cobra.Command {
	var name string
	var passphrase string
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a passphrase-protected Ed25519 key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var err error
			if passphrase == "" && !opt.JSON {
				passphrase, err = promptPassphrase(true)
				if err != nil {
					return err
				}
			}
			if passphrase == "" {
				return fmt.Errorf("a passphrase is required")
			}

			pub, priv, err := crypto.GenerateKey()
			if err != nil {
				return err
			}
			pubHex := crypto.EncodePublic(pub)
			privHex := crypto.EncodePrivate(priv)

			dir := filepath.Join(opt.Home, "keys")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			enc, err := crypto.EncryptPrivateKey(privHex, passphrase, pubHex)
			if err != nil {
				return err
			}
			path := filepath.Join(dir, name+".key")
			if err := os.WriteFile(path, enc, 0o600); err != nil {
				return err
			}
			addr := crypto.Address(pub)

			if opt.JSON {
				(&app{opt: opt}).printJSON(map[string]string{
					"name": name, "address": addr, "public": pubHex, "path": path,
				})
				return nil
			}
			fmt.Printf("Generated key %q\n", name)
			fmt.Printf("  address : %s\n", addr)
			fmt.Printf("  public  : %s\n", pubHex)
			fmt.Printf("  file    : %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "key", "key name (file: <name>.key)")
	cmd.Flags().StringVar(&passphrase, "passphrase", "", "passphrase (omit to be prompted)")
	return cmd
}

// promptPassphrase reads a passphrase from the terminal. When confirm is true
// the user is asked twice.
func promptPassphrase(confirm bool) (string, error) {
	fmt.Fprint(os.Stderr, "Passphrase: ")
	p1, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	if !confirm {
		return string(p1), nil
	}
	fmt.Fprint(os.Stderr, "Confirm passphrase: ")
	p2, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	if string(p1) != string(p2) {
		return "", fmt.Errorf("passphrases do not match")
	}
	return string(p1), nil
}
