package main

import (
	"fmt"
	"os"

	"github.com/ethanvc/study/golangproj/dkit"
	"github.com/spf13/cobra"
)

// GOOS=linux GOARCH=amd64 go build -o dkit main.go
// GOBIN=$(pwd) go install github.com/ethanvc/study/golangproj/dkit/cmd/dkit@latest
func main() {
	rootCmd := &cobra.Command{
		Use:   "dkit",
		Short: "dkit",
		Long:  `dkit`,
	}
	dkit.AddDeleteMergedBranchCmd(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
