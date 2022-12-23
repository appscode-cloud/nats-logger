package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/povsister/scp"
	"github.com/tamalsaha/ssh-exec-demo/internal/util"
	"golang.org/x/crypto/ssh"
	passgen "gomodules.xyz/password-generator"
	"gomodules.xyz/signals"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

const (
	maxRetries      int           = 1000
	backoffTimeSecs time.Duration = 10
)

func main() {
	c := NewClient()

	machineName := "capi-" + passgen.GenerateForCharset(6, passgen.AlphaNum)
	ins, err := createInstance(c, machineName, 1103682)
	if err != nil {
		panic(err)
	}
	fmt.Println("instance id:", ins.ID)

	data, err := json.MarshalIndent(ins.IPv4, "", "  ")
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

	ctx := signals.SetupSignalContext()

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

	var existingSSHClient *ssh.Client
	for i := 0; i < maxRetries; i++ {
		existingSSHClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", addr), sshConf)
		if err != nil {
			fmt.Println("wait for ssh", i)
			time.Sleep(backoffTimeSecs * time.Second)
		} else {
			err = nil
			fmt.Println("connected to ssh")
			break
		}
	}
	if err != nil {
		return err
	}

	//scpClient, err := scp.NewClient(addr, sshConf, &scp.ClientOption{})
	//if err != nil {
	//	return err
	//}
	//

	// Build a SCP client based on existing "golang.org/x/crypto/ssh.Client"
	scpClient, err := scp.NewClientFromExistingSSH(existingSSHClient, &scp.ClientOption{})
	if err != nil {
		return err
	}
	defer scpClient.Close()

	_, err = ExecuteTCPCommand(existingSSHClient, "ls -l", sshConf)
	if err != nil {
		return err
	}
	// fmt.Println(out)

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
		klog.Infoln("waiting for stacksript to finish", "attempt", attempt)

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

func ExecuteTCPCommand(conn *ssh.Client, command string, config *ssh.ClientConfig) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	session.Stdout = DefaultWriter
	session.Stderr = DefaultWriter
	session.Stdin = os.Stdin
	if config.User != "root" {
		command = fmt.Sprintf("sudo %s", command)
	}
	_ = session.Run(command)
	output := DefaultWriter.Output()
	_ = session.Close()
	return output, nil
}

var DefaultWriter = &StringWriter{
	data: make([]byte, 0),
}

type StringWriter struct {
	data []byte
}

func (s *StringWriter) Flush() {
	s.data = make([]byte, 0)
}

func (s *StringWriter) Output() string {
	return string(s.data)
}

func (s *StringWriter) Write(b []byte) (int, error) {
	klog.Infoln("$ ", string(b))
	s.data = append(s.data, b...)
	return len(b), nil
}

func consumer(ctx context.Context, subject string) error {
	addr := "this-is-nats.appscode.ninja:4222"
	nc, err := util.NewConnection(addr, "")
	if err != nil {
		return fmt.Errorf("could not connect to NATS: %s\n", err)
	}

	lines := make(chan *nats.Msg, 8*1024)

	_, err = nc.ChanSubscribe(subject, lines)
	if err != nil {
		return err
	}

	go func() {
		klog.Infof("Waiting for messages on subject %s @ %s", subject, nc.ConnectedUrl())

		for {
			select {
			case m := <-lines:
				err = handleLine(m)
				if err != nil {
					klog.Infof("Could not handle line %q: %s", string(m.Data), err)
				}

			case <-ctx.Done():
				nc.Close()
				close(lines)

				return
			}
		}
	}()

	return nil
}

// handleLine handles an individual line by parsing its subject and saving to the
// host specific log
func handleLine(m *nats.Msg) (err error) {
	_, err = fmt.Fprintln(os.Stdout, string(m.Data))
	return err
}
