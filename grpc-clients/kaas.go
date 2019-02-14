package clients

import (
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	pb "github.com/RTradeLtd/grpc/krab"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// KaasClient is how we interface with the kaas grpc key manager
type KaasClient struct {
	pb.ServiceClient
	conn *grpc.ClientConn
}

// NewKaasClient is used to instantiate our kaas client in primary or fallback mode
func NewKaasClient(opts config.Services, fallback bool) (*KaasClient, error) {
	var (
		url      string
		dialOpts []grpc.DialOption
	)
	if fallback {
		if opts.Krab.Fallback.TLS.CertPath != "" {
			creds, err := credentials.NewClientTLSFromFile(opts.Krab.Fallback.TLS.CertPath, "")
			if err != nil {
				return nil, fmt.Errorf("could not load tls cert: %s", err)
			}
			dialOpts = append(dialOpts,
				grpc.WithTransportCredentials(creds),
				grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Krab.AuthKey, true)))
		} else {
			dialOpts = append(dialOpts,
				grpc.WithInsecure(),
				grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Krab.AuthKey, false)))
		}
		url = opts.Krab.Fallback.URL
	} else {
		if opts.Krab.TLS.CertPath != "" {
			creds, err := credentials.NewClientTLSFromFile(opts.Krab.TLS.CertPath, "")
			if err != nil {
				return nil, fmt.Errorf("could not load tls cert: %s", err)
			}
			dialOpts = append(dialOpts,
				grpc.WithTransportCredentials(creds),
				grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Krab.AuthKey, true)))
		} else {
			dialOpts = append(dialOpts,
				grpc.WithInsecure(),
				grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Krab.AuthKey, false)))
		}
		url = opts.Krab.URL
	}
	conn, err := grpc.Dial(url, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &KaasClient{
		conn:          conn,
		ServiceClient: pb.NewServiceClient(conn),
	}, nil
}

// Close shuts down the client's gRPC connection
func (kc *KaasClient) Close() error { return kc.conn.Close() }
