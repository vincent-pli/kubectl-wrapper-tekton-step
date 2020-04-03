package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	manifestPath = "/tmp/manifest.yaml"
)

func init() {
}

func main() {
	var action string
	var mergeStrategy string
	var manifest string

	flag.StringVar(&action, "action", "delete", "The action on the resource.")
	flag.StringVar(&mergeStrategy, "merge-strategy", "strategic", "The merge strtegy when using action patch.")
	flag.StringVar(&manifest, "manifest", "", "The content of resource.")
	flag.Parse()

	err := ioutil.WriteFile(manifestPath, []byte(manifest), 0644)
	if err != nil {
		log.Errorf("Write manifest to file failed: %+v:", err)
		os.Exit(1)
	}

	cmd := exec.Command("/bin/sh", "/builder/kubectl.bash")
	if err != nil {
		log.Errorf("Initialize script failed: %+v:", err)
		os.Exit(1)
	}

	isDelete := action == "delete"
	args := []string{
		action,
	}
	output := "json"
	if isDelete {
		args = append(args, "--ignore-not-found")
		output = "name"
	}

	if action == "patch" {
		args = append(args, "--type")
		args = append(args, mergeStrategy)
		args = append(args, "-p")

		buff, err := ioutil.ReadFile(manifestPath)
		if err != nil {
			log.Errorf("Read menifest file failed: %v", err)
			os.Exit(1)
		}

		args = append(args, string(buff))
	}

	args = append(args, "-f")
	args = append(args, manifestPath)
	args = append(args, "-o")
	args = append(args, output)
	cmd = exec.Command("kubectl", args...)
	log.Info(strings.Join(cmd.Args, " "))
	out, err := cmd.Output()
	if err != nil {
		exErr := err.(*exec.ExitError)
		errMsg := strings.TrimSpace(string(exErr.Stderr))
		log.Errorf("Run kubectl command failed with: %v and %v", exErr, errMsg)
		os.Exit(1)
	}
	if action == "delete" {
		os.Exit(0)
	}
	obj := unstructured.Unstructured{}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		log.Errorf("Unmarshl output failed: %v", err)
		os.Exit(1)
	}
	resourceName := fmt.Sprintf("%s.%s/%s", obj.GroupVersionKind().Kind, obj.GroupVersionKind().Group, obj.GetName())
	log.Infof("%s/%s", obj.GetNamespace(), resourceName)
}
