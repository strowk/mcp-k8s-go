package tools

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const timeout = 5 * time.Second

func NewPodExecCommandTool(pool k8s.ClientPool) fxctx.Tool {
	k8sNamespace := "namespace"
	k8sPodName := "pod"
	execCommand := "command"
	k8sContext := "context"
	stdin := "stdin"
	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(k8sContext, "Kubernetes context name, defaults to current context"),
		toolinput.WithRequiredString(k8sNamespace, "Namespace where pod is located"),
		toolinput.WithRequiredString(k8sPodName, "Name of the pod to execute command in"),
		toolinput.WithRequiredString(execCommand, "Command to be executed"),
		toolinput.WithString(stdin, "Standard input to the command, defaults to empty string"),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "k8s-pod-exec",
			Description: utils.Ptr("Execute command in Kubernetes pod"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return errResponse(err)
			}
			k8sNamespace, err := input.String(k8sNamespace)
			if err != nil {
				return errResponse(fmt.Errorf("invalid input namespace: %w", err))
			}
			k8sPodName, err := input.String(k8sPodName)
			if err != nil {
				return errResponse(fmt.Errorf("invalid input pod: %w", err))
			}
			execCommand, err := input.String(execCommand)
			if err != nil {
				return errResponse(fmt.Errorf("invalid input command: %w", err))
			}
			k8sContext := input.StringOr(k8sContext, "")
			stdin := input.StringOr(stdin, "")

			kubeconfig := k8s.GetKubeConfigForContext(k8sContext)
			config, err := kubeconfig.ClientConfig()
			if err != nil {
				return errResponse(fmt.Errorf("invalid config: %w", err))
			}
			execResult, err := cmdExecuter(pool, config, k8sPodName, k8sNamespace, execCommand, k8sContext, stdin, ctx)
			if err != nil {
				return errResponse(fmt.Errorf("command execute failed: %w", err))
			}

			var content mcp.TextContent
			contents := []interface{}{}
			content, err = NewJsonContent(execResult)
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

type ExecResult struct {
	Stdout interface{} `json:"stdout"`
	Stderr interface{} `json:"stderr"`
}

func cmdExecuter(
	pool k8s.ClientPool,
	config *rest.Config,
	podName,
	namespace,
	cmd,
	k8sContext,
	stdin string,
	ctx context.Context,
) (ExecResult, error) {
	execResult := ExecResult{}
	clientset, err := pool.GetClientset(k8sContext)
	if err != nil {
		return execResult, err
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return execResult, err
	}

	if len(pod.Spec.Containers) == 0 {
		return execResult, fmt.Errorf("pod %s has no containers", podName)
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
		return execResult, err
	}

	withTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // release resources if operation finishes before timeout

	var stdout, stderr bytes.Buffer
	if err = executor.StreamWithContext(withTimeout, remotecommand.StreamOptions{
		Stdin:  strings.NewReader(stdin),
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		if err == context.DeadlineExceeded {
			return execResult, fmt.Errorf("command timed out after %s", timeout)
		}
		return execResult, err
	}
	execResult.Stdout = stdout.String()
	execResult.Stderr = stderr.String()
	return execResult, nil
}
