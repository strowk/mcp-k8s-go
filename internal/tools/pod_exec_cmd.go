package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

func NewPodExecCommandTool(pool k8s.ClientPool) fxctx.Tool {
	k8sNamespace := "namespace"
	k8sPodName := "podName"
	execCommand := "command"
	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(k8sNamespace, "The name of the namespace where the pod to execute the command is located."),
		toolinput.WithString(k8sPodName, "The name of the pod in which the command needs to be executed."),
		toolinput.WithString(execCommand, "The command to be executed inside the pod."),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "pod-exec-command",
			Description: utils.Ptr("Enter the specified namespace's pod, execute the specified command, and return the execution result."),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return errResponse(err)
			}
			k8sNamespace := input.StringOr(k8sNamespace, "")
			k8sPodName := input.StringOr(k8sPodName, "")
			execCommand := input.StringOr(execCommand, "")

			config := getConfigFromKubeFile(os.Getenv("KUBECONFIG"))
			output, err := cmdExecuter(config, k8sPodName, k8sNamespace, execCommand)
			if err != nil {
				return errResponse(err)
			}

			var content mcp.TextContent
			contents := []interface{}{}
			content, err = NewJsonContent(ExecResult{
				ExecResultKey:   "podName",
				ExecResultValue: output[0],
			})
			if err != nil {
				return errResponse(err)
			}
			contents = append(contents, content)
			content, err = NewJsonContent(ExecResult{
				ExecResultKey:   "command",
				ExecResultValue: output[1],
			})
			if err != nil {
				return errResponse(err)
			}
			contents = append(contents, content)
			content, err = NewJsonContent(ExecResult{
				ExecResultKey:   "stdout",
				ExecResultValue: output[2],
			})
			if err != nil {
				return errResponse(err)
			}
			contents = append(contents, content)
			content, err = NewJsonContent(ExecResult{
				ExecResultKey:   "stdin",
				ExecResultValue: output[3],
			})
			if err != nil {
				return errResponse(err)
			}
			contents = append(contents, content)

			return &mcp.CallToolResult{
				Meta:    map[string]interface{}{},
				Content: contents,
				IsError: utils.Ptr(false),
			}
		},
	)
}

// ExecResult
type ExecResult struct {
	ExecResultKey   string      `json:"execResultKey"`
	ExecResultValue interface{} `json:"execResultValue"`
}

func cmdExecuter(config *rest.Config, podName, namespace, cmd string) ([]interface{}, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(pod.Spec.Containers) == 0 {
		return nil, fmt.Errorf("Pod %s has no containers", podName)
	}

	containerName := pod.Spec.Containers[0].Name
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"sh", "-c", cmd},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return nil, err
	}
	var stdout, stderr bytes.Buffer
	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return nil, err
	}
	ret := []interface{}{podName, cmd, stdout.String(), stderr.String()}
	return ret, nil
}

func getConfigFromKubeFile(kubeConfigFilePath string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
	config.TLSClientConfig.Insecure = true
	config.Insecure = true
	if err != nil {
		panic(err)
	}
	return config
}
