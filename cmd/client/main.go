package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	ManifestPath = "/tmp/manifest.yaml"
	Separator    = ","
	OutputFile   = "/tekton/output"
)

func init() {
}

func main() {
	var action string
	var mergeStrategy string
	var manifest string
	var successCondition string
	var failureCondition string
	var output string

	flag.StringVar(&action, "action", "delete", "The action on the resource.")
	flag.StringVar(&mergeStrategy, "merge-strategy", "strategic", "The merge strtegy when using action patch.")
	flag.StringVar(&manifest, "manifest", "", "The content of resource.")
	flag.StringVar(&successCondition, "success-condition", "", "A label selector express to decide if the action on resource is success.")
	flag.StringVar(&failureCondition, "failure-condition", "", "A label selector express to decide if the action on resource is failure.")
	flag.StringVar(&output, "output", "", "An express to retrieval data from resource.")

	flag.Parse()

	err := ioutil.WriteFile(ManifestPath, []byte(manifest), 0644)
	if err != nil {
		log.Errorf("Write manifest to file failed: %+v:", err)
		os.Exit(1)
	}

	cmd := exec.Command("/bin/sh", "/builder/kubectl.bash")
	_, err = cmd.Output()
	if err != nil {
		log.Errorf("Initialize script failed: %+v:", err)
	}

	isDelete := action == "delete"
	resourceNamespace, resourceName, err := execResource(action, mergeStrategy)
	if err != nil {
		log.Errorf("Execute resource failed: %+v:", err)
		os.Exit(1)
	}

	if !isDelete {
		err = waitResource(resourceNamespace, resourceName, successCondition, failureCondition)
		if err != nil {
			log.Errorf("Waiting resource failed: %+v:", err)
			os.Exit(1)
		}

		err = saveResult(resourceNamespace, resourceName, output)
		if err != nil {
			log.Errorf("Write output failed: %+v:", err)
			os.Exit(1)
		}
	}

}

func execResource(action, mergeStrategy string) (string, string, error) {
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

		buff, err := ioutil.ReadFile(ManifestPath)
		if err != nil {
			log.Errorf("Read menifest file failed: %v", err)
			return "", "", err
		}

		args = append(args, string(buff))
	}

	args = append(args, "-f")
	args = append(args, ManifestPath)
	args = append(args, "-o")
	args = append(args, output)
	cmd := exec.Command("kubectl", args...)
	log.Info(strings.Join(cmd.Args, " "))
	out, err := cmd.Output()
	if err != nil {
		exErr := err.(*exec.ExitError)
		errMsg := strings.TrimSpace(string(exErr.Stderr))
		log.Errorf("Run kubectl command failed with: %v and %v", exErr, errMsg)
		return "", "", err
	}
	if action == "delete" {
		return "", "", nil
	}
	obj := unstructured.Unstructured{}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		log.Errorf("Unmarshl output failed: %v", err)
		return "", "", err
	}
	resourceName := fmt.Sprintf("%s.%s/%s", obj.GroupVersionKind().Kind, obj.GroupVersionKind().Group, obj.GetName())
	log.Infof("%s/%s", obj.GetNamespace(), resourceName)

	return obj.GetNamespace(), resourceName, nil
}

func waitResource(namespace, name, successCondition, failureCondition string) error {
	if successCondition == "" && failureCondition == "" {
		return nil
	}
	var successReqs labels.Requirements
	if successCondition != "" {
		successSelector, err := labels.Parse(successCondition)
		if err != nil {
			return err
		}
		log.Infof("Waiting for conditions: %s", successSelector)
		successReqs, _ = successSelector.Requirements()
	}

	var failReqs labels.Requirements
	if failureCondition != "" {
		failSelector, err := labels.Parse(failureCondition)
		if err != nil {
			return err
		}
		log.Infof("Failing for conditions: %s", failSelector)
		failReqs, _ = failSelector.Requirements()
	}

	// Start the condition result reader using PollImmediateInfinite
	// Poll intervall of 5 seconds serves as a backoff intervall in case of immediate result reader failure
	err := wait.PollImmediateInfinite(time.Second*5,
		func() (bool, error) {
			isErrRetry, err := checkResourceState(namespace, name, successReqs, failReqs)

			if err == nil {
				log.Infof("Returning from successful wait for resource %s", name)
				return true, nil
			}

			if isErrRetry {
				log.Infof("Waiting for resource %s resulted in retryable error %v", name, err)
				return false, nil
			}

			log.Warnf("Waiting for resource %s resulted in non-retryable error %v", name, err)
			return false, err
		})

	if err != nil {
		if err == wait.ErrWaitTimeout {
			log.Warnf("Waiting for resource %s resulted in timeout due to repeated errors", name)
		} else {
			log.Warnf("Waiting for resource %s resulted in error %v", name, err)
		}
		return err
	}

	return nil
}

func checkIfResourceDeleted(resourceName string, resourceNamespace string) bool {
	args := []string{"get", resourceName}
	if resourceNamespace != "" {
		args = append(args, "-n", resourceNamespace)
	}
	cmd := exec.Command("kubectl", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if strings.Contains(stderr.String(), "NotFound") {
			return true
		}
		log.Warnf("Got error %v when checking if the resource %s in namespace %s is deleted", err, resourceName, resourceNamespace)
		return false
	}
	return false
}

// Function to do the kubectl get -w command and then waiting on json reading.
func checkResourceState(resourceNamespace string, resourceName string, successReqs labels.Requirements, failReqs labels.Requirements) (bool, error) {
	cmd, reader, err := startKubectlWaitCmd(resourceNamespace, resourceName)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	for {
		jsonBytes, err := readJSON(reader)

		if err != nil {
			resultErr := err
			log.Warnf("Json reader returned error %v. Calling kill (usually superfluous)", err)
			// We don't want to write OS specific code so we don't want to call syscall package code. But that means
			// there is no way to figure out if a process is running or not in an asynchronous manner. exec.Wait will
			// always block and we need to call that to get the exit code of the process. So we will unconditionally
			// call exec.Process.Kill and then assume that wait will not block after that. Two things may happen:
			// 1. Process already exited and kill does nothing (returns error which we ignore) and then we call
			//    Wait and get the proper return value
			// 2. Process is running gets, killed with exec.Process.Kill call and Wait returns an error code and we give up
			//    and don't retry
			_ = cmd.Process.Kill()

			log.Warnf("Command for kubectl get -w for %s exited. Getting return value using Wait", resourceName)
			err = cmd.Wait()
			if err != nil {
				log.Warnf("cmd.Wait for kubectl get -w command for resource %s returned error %v",
					resourceName, err)
				resultErr = err
			} else {
				log.Infof("readJSon failed for resource %s but cmd.Wait for kubectl get -w command did not error", resourceName)
			}
			return true, resultErr
		}

		if checkIfResourceDeleted(resourceName, resourceNamespace) {
			return false, err
		}

		log.Info(string(jsonBytes))
		ls := gjsonLabels{json: jsonBytes}
		for _, req := range failReqs {
			failed := req.Matches(ls)
			msg := fmt.Sprintf("failure condition '%s' evaluated %v", req, failed)
			log.Infof(msg)
			if failed {
				// TODO: need a better error code instead of BadRequest
				return false, fmt.Errorf("Action failed: %s/%s", resourceNamespace, resourceName)
			}
		}
		numMatched := 0
		for _, req := range successReqs {
			matched := req.Matches(ls)
			log.Infof("success condition '%s' evaluated %v", req, matched)
			if matched {
				numMatched++
			}
		}
		log.Infof("%d/%d success conditions matched", numMatched, len(successReqs))
		if numMatched >= len(successReqs) {
			return false, nil
		}
	}
}

// Start Kubectl command Get with -w return error if unable to start command
func startKubectlWaitCmd(resourceNamespace string, resourceName string) (*exec.Cmd, *bufio.Reader, error) {
	args := []string{"get", resourceName, "-w", "-o", "json"}
	if resourceNamespace != "" {
		args = append(args, "-n", resourceNamespace)
	}
	cmd := exec.Command("kubectl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	reader := bufio.NewReader(stdout)
	log.Info(strings.Join(cmd.Args, " "))
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	return cmd, reader, nil
}

// readJSON reads from a reader line-by-line until it reaches "}\n" indicating end of json
func readJSON(reader *bufio.Reader) ([]byte, error) {
	var buffer bytes.Buffer
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		isDelimiter := len(line) == 2 && line[0] == byte('}')
		line = bytes.TrimSpace(line)
		_, err = buffer.Write(line)
		if err != nil {
			return nil, err
		}
		if isDelimiter {
			break
		}
	}
	return buffer.Bytes(), nil
}

// gjsonLabels is an implementation of labels.Labels interface
// which allows us to take advantage of k8s labels library
// for the purposes of evaluating fail and success conditions
type gjsonLabels struct {
	json []byte
}

// Has returns whether the provided label exists.
func (g gjsonLabels) Has(label string) bool {
	return gjson.GetBytes(g.json, label).Exists()
}

// Get returns the value for the provided label.
func (g gjsonLabels) Get(label string) string {
	return gjson.GetBytes(g.json, label).String()
}

type outputItem struct {
	name string
	value string
}

// Save result to files
func saveResult(resourceNamespace, resourceName, output string) error {
	outputs := []outputItem{}

	if len(output) == 0 {
		log.Infof("No output parameters")
		return nil
	}

	log.Infof("Saving resource output parameters")
	for _, param := range strings.Split(output, Separator) {
		param = strings.Trim(param, " ")
		if len(param) == 0 {
			continue
		}
		var cmd *exec.Cmd
		args := []string{"get", resourceName, "-o", fmt.Sprintf("jsonpath=%s", param)}
		if resourceNamespace != "" {
			args = append(args, "-n", resourceNamespace)
		}
		cmd = exec.Command("kubectl", args...)

		log.Info(cmd.Args)
		out, err := cmd.Output()
		if err != nil {
			log.Infof("Retrieval output failed %s/%s with error: %+v", resourceNamespace, resourceName, err)
			return err
		}
		ot := outputItem{}
		ot.name = param
		ot.value = string(out)
		outputs = append(outputs, ot)
		log.Infof("Saved output parameter: %s, value: %s", param, output)
	}

	err := writeFiles(outputs)
	if err != nil {
		return err
	}

	return nil
}

func writeFiles(outputs []outputItem) error {
	outputBytes, err := json.Marshal(outputs)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(OutputFile, outputBytes, 0644)
	if err != nil {
		log.Errorf("Write output to file failed: %+v:", err)
		return err
	}
	return nil
}
