package dkit

import "github.com/spf13/cobra"

func AddGrpcCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use: "grpc",
	}
	host := cmd.Flags().String("host", "127.0.0.1:8888", "server instance address")
	body := cmd.Flags().String("body", "", "reqeust content")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return SendGrpcRequest(&SendGrpcRequestReq{
			Host: *host,
			Body: *body,
		})
	}
	rootCmd.AddCommand(cmd)
}

type SendGrpcRequestReq struct {
	Host string
	Body string
}

func SendGrpcRequest(req *SendGrpcRequestReq) error {
	return nil
}
