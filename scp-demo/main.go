package main

import (
	"context"
	"os"
	"time"

	"github.com/povsister/scp"
)

func main() {
	if err := runSCP(); err != nil {
		panic(err)
	}
}

func runSCP() error {
	// Build a SSH config from username/password
	// sshConf := scp.NewSSHConfigFromPassword("username", "password")

	// Build a SSH config from private key
	privPEM, err := os.ReadFile("/Users/tamal/.ssh/id_rsa") // "/path/to/privateKey"
	if err != nil {
		return err
	}
	// without passphrase
	sshConf, err := scp.NewSSHConfigFromPrivateKey("root", privPEM)
	if err != nil {
		return err
	}
	// with passphrase
	// sshConf, err := scp.NewSSHConfigFromPrivateKey("username", privPEM, passphrase)

	// Dial SSH to "my.server.com:22".
	// If your SSH server does not listen on 22, simply suffix the address with port.
	// e.g: "my.server.com:1234"
	scpClient, err := scp.NewClient("139.144.38.197", sshConf, &scp.ClientOption{})

	// Build a SCP client based on existing "golang.org/x/crypto/ssh.Client"
	// scpClient, err := scp.NewClientFromExistingSSH(existingSSHClient, &scp.ClientOption{})

	defer scpClient.Close()

	// Do the file transfer without timeout/context
	err = scpClient.CopyFileToRemote("/path/to/local/file", "/path/at/remote", &scp.FileTransferOption{})

	// Do the file copy with timeout, context and file properties preserved.
	// Note that the context and timeout will both take effect.
	fo := &scp.FileTransferOption{
		Context:      context.TODO(),
		Timeout:      30 * time.Second,
		PreserveProp: true,
	}
	err = scpClient.CopyFileFromRemote("/root/success2.txt", "/tmp/success.txt", fo)
	if err != nil {
		return err
	}
	return nil

	// scp: /root/success2.txt: No such file or directory
}
