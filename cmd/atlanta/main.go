package main

import (
	"fmt"
	"github.com/rmohr/atlanta/cmd/atlanta/node"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "atlanta",
		Short: "atlante helps getting out best performance for your KubeVirt VMs",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rootCmd.AddCommand(node.NewCmdNode(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}))
	rootCmd.Execute()
}
