package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"

	// "os"
	"os/exec"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	// "k8s.io/apimachinery/pkg/runtime"
)

// var (
// setupLog = ctrl.Log.WithName("setup")
// )

func init() {
}

func main() {
	fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	rand.Seed(time.Now().UTC().UnixNano())
	var action string
	var mergeStrategy string
	var manifest string

	flag.StringVar(&action, "action", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&mergeStrategy, "merge-strategy", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&manifest, "manifest", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()

	args := []string{
		"get",
		"pod",
		"test",
		"-ojson",
	}
	cmd := exec.Command("/bin/sh", "/builder/kubectl.bash")
	time.Sleep(60 * time.Second)
	cmd = exec.Command("kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("hekko, %+v", err)
	}

	obj := unstructured.Unstructured{}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		fmt.Printf("hekko1, %+v", err)
	}
	fmt.Printf("------- %+v", obj)
	fmt.Printf("+++++++ %+v", out)
	// isDelete := action == "delete"
	// args := []string{
	// 	action,
	// }
	// output := "json"
	// if isDelete {
	// 	args = append(args, "--ignore-not-found")
	// 	output = "name"
	// }

	// if action == "patch" {
	// 	mergeStrategy := "strategic"
	// 	if we.Template.Resource.MergeStrategy != "" {
	// 		mergeStrategy = we.Template.Resource.MergeStrategy
	// 	}

	// 	args = append(args, "--type")
	// 	args = append(args, mergeStrategy)

	// 	args = append(args, "-p")
	// 	buff, err := ioutil.ReadFile(manifestPath)

	// 	if err != nil {
	// 		return "", "", errors.New(errors.CodeBadRequest, err.Error())
	// 	}

	// 	args = append(args, string(buff))
	// }

	// args = append(args, "-f")
	// args = append(args, manifestPath)
	// args = append(args, "-o")
	// args = append(args, output)
	// cmd := exec.Command("kubectl", args...)
	// log.Info(strings.Join(cmd.Args, " "))
	// out, err := cmd.Output()
	// if err != nil {
	// 	exErr := err.(*exec.ExitError)
	// 	errMsg := strings.TrimSpace(string(exErr.Stderr))
	// 	return "", "", errors.New(errors.CodeBadRequest, errMsg)
	// }
	// if action == "delete" {
	// 	return "", "", nil
	// }
	// obj := unstructured.Unstructured{}
	// err = json.Unmarshal(out, &obj)
	// if err != nil {
	// 	return "", "", err
	// }
	// resourceName := fmt.Sprintf("%s.%s/%s", obj.GroupVersionKind().Kind, obj.GroupVersionKind().Group, obj.GetName())
	// log.Infof("%s/%s", obj.GetNamespace(), resourceName)
	// cmd := plugins.Register()
	// if err := cmd.Execute(); err != nil {
	// 	_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	// 	os.Exit(1)
	// }
}
