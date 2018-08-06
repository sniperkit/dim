package cmd

import (
	"context"
	"fmt"

	// external
	"github.com/spf13/cobra"

	// internal
	"github.com/sniperkit/dim/pkg/cli"
	"github.com/sniperkit/dim/pkg/core/utils"
)

func newGenPasswdCommand(c *cli.Cli, rootCommand *cobra.Command, ctx context.Context) {
	genpasswdCommand := &cobra.Command{
		Use:   "genpasswd QUERY",
		Short: "Encode a password in sha256",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenPasswd(c, args)
		},
	}

	rootCommand.AddCommand(genpasswdCommand)
}

func runGenPasswd(c *cli.Cli, args []string) error {
	var password string
	if len(args) > 0 {
		password = args[0]
	} else {
		for password == "" {
			fmt.Fprint(c.Out, "Password :")
			cli.ReadPassword(&password)
		}
	}
	fmt.Fprintln(c.Out, utils.Sha256(password))
	return nil
}
