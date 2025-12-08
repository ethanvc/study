package main

import (
	"fmt"
	"os"

	"github.com/ethanvc/study/golangproj/dkit"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dkit",
		Short: "dkit",
		Long:  `dkit`,
	}
	dkit.AddGrpcCmd(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
