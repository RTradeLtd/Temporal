package temporal

import (
	context "context"
	"io"
	"log"

	grpc "google.golang.org/grpc"

	"github.com/RTradeLtd/sdk/go/temporal/store"
)

// UploadFile is a convenience function for uploading a single file
func UploadFile(
	ctx context.Context,
	client store.TemporalStoreClient,
	file io.Reader,
	holdTime int32,
	fileOpts *store.ObjectOptions,
	grpcOpts ...grpc.CallOption,
) (*store.Object, error) {
	stream, err := client.Upload(ctx, grpcOpts...)
	if err != nil {
		log.Fatal(err)
	}

	// declare file options
	stream.Send(&store.UploadReq{HoldTime: holdTime, Options: fileOpts})

	// upload file - chunked at 5mb each
	buf := make([]byte, 5e+6)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		stream.Send(&store.UploadReq{Blob: &store.Blob{Content: buf[:n]}})
	}

	// done
	return stream.CloseAndRecv()
}

// DownloadFile is a convenience function for downloading a single file
func DownloadFile(
	ctx context.Context,
	client store.TemporalStoreClient,
	file io.Writer,
	download *store.DownloadReq,
	grpcOpts ...grpc.CallOption,
) error {
	stream, err := client.Download(ctx, download, grpcOpts...)
	if err != nil {
		return err
	}
	for {
		b, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if _, err := file.Write(b.Content); err != nil {
			return err
		}
	}
	return stream.CloseSend()
}
