package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	passgen "gomodules.xyz/password-generator"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/povsister/scp"
)

func main() {
	c := NewClient()

	machineName := "gh-runner-" + passgen.GenerateForCharset(6, passgen.AlphaNum)
	ins, err := createInstance(c, machineName, 1103682)
	if err != nil {
		panic(err)
	}
	fmt.Println("instance id:", ins.ID)

	data, err := json.MarshalIndent(ins, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	var addr string
	for _, ip := range ins.IPv4 {
		if !(ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsLoopback()) {
			addr = ip.String()
			break
		}
	}
	if addr == "" {
		panic(fmt.Errorf("failed to detect IP for Linode instance id: %s", ins.ID))
	}
	if err := runSCP(addr, "/Users/tamal/.ssh/id_rsa", "root"); err != nil {
		panic(err)
	}
}

func main_scp() {
	if err := runSCP("139.144.38.197", "/Users/tamal/.ssh/id_rsa", "root"); err != nil {
		panic(err)
	}
}

func runSCP(addr, privateKey, username string) error {
	// Build a SSH config from username/password
	// sshConf := scp.NewSSHConfigFromPassword("username", "password")

	// Build a SSH config from private key
	privPEM, err := os.ReadFile(privateKey) // "/path/to/privateKey"
	if err != nil {
		return err
	}
	// without passphrase
	sshConf, err := scp.NewSSHConfigFromPrivateKey(username, privPEM)
	if err != nil {
		return err
	}
	// with passphrase
	// sshConf, err := scp.NewSSHConfigFromPrivateKey("username", privPEM, passphrase)

	// Dial SSH to "my.server.com:22".
	// If your SSH server does not listen on 22, simply suffix the address with port.
	// e.g: "my.server.com:1234"
	scpClient, err := scp.NewClient(addr, sshConf, &scp.ClientOption{})
	if err != nil {
		return err
	}
	// Build a SCP client based on existing "golang.org/x/crypto/ssh.Client"
	// scpClient, err := scp.NewClientFromExistingSSH(existingSSHClient, &scp.ClientOption{})

	defer scpClient.Close()

	//// Do the file transfer without timeout/context
	//err = scpClient.CopyFileToRemote("/path/to/local/file", "/path/at/remote", &scp.FileTransferOption{})

	// Do the file copy with timeout, context and file properties preserved.
	// Note that the context and timeout will both take effect.

	err = waitUntilScriptDone(scpClient)
	if err != nil {
		return err
	}

	fo := &scp.FileTransferOption{
		Context:      context.TODO(),
		Timeout:      30 * time.Second,
		PreserveProp: true,
	}
	err = scpClient.CopyFileFromRemote("/root/stackscript.log", "/tmp/stackscript.log", fo)
	if err != nil {
		return err
	}

	return nil

	// scp: /root/success2.txt: No such file or directory
}

func waitUntilScriptDone(scpClient *scp.Client) error {
	attempt := 0
	klog.Infoln("waiting for stackscript to complete")

	fo := &scp.FileTransferOption{
		Context:      context.TODO(),
		Timeout:      30 * time.Second,
		PreserveProp: true,
	}
	return wait.PollImmediate(RetryInterval, RetryTimeout, func() (bool, error) {
		attempt++

		err := scpClient.CopyFileFromRemote("/root/success.txt", "/tmp/success.txt", fo)
		if err != nil {
			if strings.Contains(err.Error(), "No such file or directory") {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
}
