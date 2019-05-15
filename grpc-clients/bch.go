package clients

import (
	"fmt"

	"github.com/RTradeLtd/config/v2"
	pb "github.com/gcash/bchwallet/rpc/walletrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// BchWalletClient provides a gRPC API
// to manage a BCH wallet
type BchWalletClient struct {
	pb.WalletServiceClient
}

// NewBchWalletClient is used to instantaite our connection
// to a bchwallet daemon for wallet management functionality
func NewBchWalletClient(opts config.Services) (*BchWalletClient, error) {
	dialOpts := make([]grpc.DialOption, 0)
	if opts.BchGRPC.Wallet.CertFile != "" {
		creds, err := credentials.NewClientTLSFromFile(opts.BchGRPC.Wallet.CertFile, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	gConn, err := grpc.Dial(opts.BchGRPC.Wallet.URL, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &BchWalletClient{pb.NewWalletServiceClient(gConn)}, nil
}
