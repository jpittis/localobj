package localobj

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	defaultMinIOUser     = "minioadmin"
	defaultMinIOPassword = "minioadmin"
	defaultMinIORegion   = "us-east-1"
)

func getAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func defaultMinIOCmdContextWithPort(port int) CmdContext {
	return func(ctx context.Context) (*exec.Cmd, error) {
		tempDir, err := os.MkdirTemp("", "minio-data-*")
		if err != nil {
			return nil, err
		}

		cmd := exec.CommandContext(ctx, "minio", "server", "--address", fmt.Sprintf(":%d", port), tempDir)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("MINIO_ROOT_USER=%s", defaultMinIOUser),
			fmt.Sprintf("MINIO_ROOT_PASSWORD=%s", defaultMinIOPassword),
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd, nil
	}
}

func defaultMinIOClientOptions(port int) func(*s3.Options) {
	return func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("http://localhost:%d", port))
		o.UsePathStyle = true
		o.Credentials = credentials.NewStaticCredentialsProvider(
			defaultMinIOUser,
			defaultMinIOPassword,
			"",
		)
		o.Region = defaultMinIORegion
	}
}
