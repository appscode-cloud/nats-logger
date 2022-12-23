package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/linode/linodego"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	passgen "gomodules.xyz/password-generator"
	"gomodules.xyz/pointer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	RetryInterval = 30 * time.Second
	RetryTimeout  = 3 * time.Minute
)

func stopRunner(e *github.WorkflowJobEvent) {
	c := NewClient()

	machineName := fmt.Sprintf("%s-%s-%d", e.Org.GetLogin(), e.Repo.GetName(), e.GetWorkflowJob().GetID())

	filter := fmt.Sprintf(`{"label" : "%v"}`, machineName)
	listOpts := &linodego.ListOptions{PageOptions: nil, Filter: filter}

	instances, err := c.ListInstances(context.Background(), listOpts)
	if err != nil {
		panic(err)
	}
	if len(instances) > 1 {
		klog.Errorf("multiple linodes found with label %v", machineName)
		return
	} else if len(instances) == 0 {
		klog.Errorf("no linode found with label %v", machineName)
		return
	}

	id := instances[0].ID
	err = c.DeleteInstance(context.Background(), id)
	if err != nil {
		panic(err)
	}
	fmt.Println("instance id:", id)
}

func startRunner(e *github.WorkflowJobEvent) {
	c := NewClient()

	machineName := fmt.Sprintf("%s-%s-%d", e.Org.GetLogin(), e.Repo.GetName(), e.GetWorkflowJob().GetID())
	fmt.Println(machineName)

	// machineName := "gh-runner-" + passgen.Generate(6)
	id, err := createInstance(c, machineName, 1018111)
	if err != nil {
		panic(err)
	}
	fmt.Println("instance id:", id)
}

func main___() {
	c := NewClient()

	machineName := "gh-runner-" + passgen.Generate(6)
	id, err := createInstance(c, machineName, 1018111)
	if err != nil {
		panic(err)
	}
	fmt.Println("instance id:", id)

	/*
		// linode/ubuntu16.04lts Ubuntu 16.04 LTS
		// linode/ubuntu18.04 Ubuntu 18.04 LTS
		// linode/ubuntu20.04 Ubuntu 20.04 LTS
		// linode/ubuntu21.10 Ubuntu 21.10
		// linode/ubuntu22.04 Ubuntu 22.04 LTS
		images, err := c.ListImages(context.Background(), &linodego.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, r := range images {
			fmt.Println(r.ID, r.Label)
		}
		fmt.Println("----------------")
	*/

	/*
		sshKeys, err := c.ListSSHKeys(context.Background(), &linodego.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, r := range sshKeys {
			fmt.Println(r.ID, r.Label)
		}
		fmt.Println("----------------")
	*/

	/*
		// ap-west in
		// ca-central ca
		// ap-southeast au
		// us-central us
		// us-west us
		// us-southeast us
		// us-east us
		// eu-west uk
		// ap-south sg
		// eu-central de
		// ap-northeast jp
		regions, err := c.ListRegions(context.Background(), &linodego.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, r := range regions {
			fmt.Println(r.ID, r.Country)
		}
		fmt.Println("----------------")
	*/

	/*
		// g6-nanode-1 Nanode 1GB
		// g6-standard-1 Linode 2GB
		// g6-standard-2 Linode 4GB
		// g6-standard-4 Linode 8GB
		// g6-standard-6 Linode 16GB
		// g6-standard-8 Linode 32GB
		// g6-standard-16 Linode 64GB
		// g6-standard-20 Linode 96GB
		// g6-standard-24 Linode 128GB
		// g6-standard-32 Linode 192GB
		// g6-dedicated-2 Dedicated 4GB
		// g6-dedicated-4 Dedicated 8GB
		// g6-dedicated-8 Dedicated 16GB
		// g6-dedicated-16 Dedicated 32GB
		// g6-dedicated-32 Dedicated 64GB
		linodeTypes, err := c.ListTypes(context.Background(), &linodego.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, r := range linodeTypes {
			fmt.Println(r.ID, r.Label)
		}
	*/

	/*
		scriptID, err := getStartupScriptID(&c)
		if err != nil {
			panic(err)
		}
		fmt.Println(scriptID)
	*/
	// scriptID := 1018111
}

func NewClient() *linodego.Client {
	token := os.Getenv("LINODE_CLI_TOKEN")
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	c := linodego.NewClient(oauth2Client)
	return &c
}

func getStartupScriptID(c *linodego.Client) (int, error) {
	scriptName := "gh-runner"
	filter := fmt.Sprintf(`{"label" : "%v"}`, scriptName)
	listOpts := &linodego.ListOptions{PageOptions: nil, Filter: filter}

	scripts, err := c.ListStackscripts(context.Background(), listOpts)
	if err != nil {
		return 0, err
	}

	if len(scripts) > 1 {
		return 0, errors.Errorf("multiple stackscript found with label %v", scriptName)
	} else if len(scripts) == 0 {
		return 0, errors.Errorf("no stackscript found with label %v", scriptName)
	}
	return scripts[0].ID, nil
}

func createInstance(c *linodego.Client, machineName string, scriptID int) (*linodego.Instance, error) {
	sshKeys, err := c.ListSSHKeys(context.Background(), &linodego.ListOptions{})
	if err != nil {
		return nil, err
	}
	authorizedKeys := make([]string, 0, len(sshKeys))
	for _, r := range sshKeys {
		authorizedKeys = append(authorizedKeys, r.SSHKey)
	}

	rootPassword := passgen.Generate(20)
	fmt.Println("rootPassword:", rootPassword)
	createOpts := linodego.InstanceCreateOptions{
		Label:          machineName,
		Region:         "us-central",
		Type:           "g6-nanode-1", //  "g6-standard-1", // "g6-nanode-1",
		RootPass:       rootPassword,
		AuthorizedKeys: authorizedKeys,
		StackScriptData: map[string]string{
			"my_var": machineName,
		},
		StackScriptID:  scriptID,
		Image:          "linode/ubuntu22.04",
		BackupsEnabled: false,
		PrivateIP:      true,
		SwapSize:       pointer.IntP(0),
	}

	instance, err := c.CreateInstance(context.Background(), createOpts)
	if err != nil {
		return nil, err
	}

	if err := waitForStatus(c, instance.ID, linodego.InstanceRunning); err != nil {
		return nil, err
	}

	return instance, nil
}

func waitForStatus(c *linodego.Client, id int, status linodego.InstanceStatus) error {
	attempt := 0
	klog.Infoln("waiting for instance status", "status", status)
	return wait.PollImmediate(RetryInterval, RetryTimeout, func() (bool, error) {
		attempt++

		instance, err := c.GetInstance(context.Background(), id)
		if err != nil {
			return false, nil
		}
		if instance == nil {
			return false, nil
		}
		klog.Infoln("current instance state", "instance", instance.Label, "status", instance.Status, "attempt", attempt)
		if instance.Status == status {
			klog.Infoln("current instance status", "status", status)
			return true, nil
		}
		return false, nil
	})
}
