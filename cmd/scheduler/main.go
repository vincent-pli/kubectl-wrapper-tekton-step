package main

import (
	"flag"
	"fmt"
	"math/rand"
	// "os"
	"time"
	"os/exec"

	// "k8s.io/apimachinery/pkg/runtime"
)

// var (
	// setupLog = ctrl.Log.WithName("setup")
// )

func init() {
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	var action string
	var mergeStrategy string
	var manifest string

	flag.StringVar(&action, "action", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&mergeStrategy, "merge-strategy", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&manifest, "manifest", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()

	args := []string{
		"delete",
		"pod",
		"test",
	}
	cmd := exec.Command("kubectl", args...)

	_, err := cmd.Output()
	if err != nil {
		fmt.Println("hekko")
	}

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
