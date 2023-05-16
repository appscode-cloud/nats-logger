/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/nats-io/nats.go"
	"github.com/rs/xid"
	"go.bytebuilders.dev/nats-logger/internal/util"
	"k8s.io/klog/v2"
)

func main() {
	source := os.Getenv("SHIPPER_FILE")
	if source == "" {
		log.Fatalf("Please set a file to publish using SHIPPER_FILE\n")
	}

	subject := os.Getenv("SHIPPER_SUBJECT")
	if subject == "" {
		log.Fatalf("Please set a NATS subject to publish to using SHIPPER_SUBJECT\n")
	}

	addr := os.Getenv("NATS_SERVER")
	creds := os.Getenv("NATS_CREDS")
	credFile, err := os.CreateTemp("", "nats-*.creds")
	if err != nil {
		log.Fatalf(err.Error())
	}
	_, err = credFile.Write([]byte(creds))
	if err != nil {
		log.Fatalf("Could not write creds: %s\n", err)
	}
	defer os.Remove(credFile.Name())
	//partition, name, err := util.Partition()
	//if err != nil {
	//	log.Fatal(err.Error())
	//}

	// addr := "this-is-nats.appscode.ninja:4222"
	nc, err := util.NewConnection(addr, credFile.Name())
	if err != nil {
		log.Fatalf("Could not connect to NATS: %s\n", err)
	}

	id := xid.New().String()
	title, ok := os.LookupEnv("SHIPPER_TITLE")
	if !ok {
		title = "Cluster Provisioning Logs"
	}

	msg := newResponse(TaskStatusStarted, id, title, "Creating Linode Instance")
	if err = nc.Publish(subject, msg); err != nil {
		log.Printf("Could not publish response")
	}

	publishClusterProvisioningLogs(source, subject, id, nc)
}

func publishClusterProvisioningLogs(source, subject, id string, nc *nats.Conn) {
	for {
		if err := publishFile(source, subject, id, nc); err != nil {
			log.Printf("Could not publish file: %s", err)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func publishFile(source, subject, id string, nc *nats.Conn) error {
	t, err := tail.TailFile(source, tail.Config{Follow: true})
	if err != nil {
		return err
	}

	log.Printf("Publishing lines from %s to %s", source, subject)

	status := TaskStatus(TaskStatusRunning)
	for line := range t.Lines {
		if status == TaskStatusRunning {
			status = generateTaskStatus(line.Text)
		}
		msg := newResponse(status, id, "", line.Text)
		if err = nc.Publish(subject, msg); err != nil {
			klog.ErrorS(err, "failed to publish log")
		}
	}

	return nil
}

func generateTaskStatus(msg string) TaskStatus {
	if strings.Contains(msg, "Cluster provision: Task failed !") {
		return TaskStatusFailed
	} else if strings.Contains(msg, "Cluster provision: Task completed successfully !") {
		return TaskStatusSuccess
	}

	return TaskStatusRunning
}

type TaskStatus string

const (
	TaskStatusPending = "Pending"
	TaskStatusStarted = "Started"
	TaskStatusRunning = "Running"
	TaskStatusFailed  = "Failed"
	TaskStatusSuccess = "Success"
)

func newResponse(status TaskStatus, id, title, msg string) []byte {
	m := map[string]string{
		"status": string(status),
		"msg":    msg,
	}
	if id != "" {
		m["id"] = id
	}
	if title != "" {
		m["step"] = title
	}
	data, _ := json.Marshal(m)
	return data
}
