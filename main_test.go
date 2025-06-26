package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/strowk/foxy-contexts/pkg/foxytest"
	"github.com/strowk/mcp-k8s-go/internal/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListContexts(t *testing.T) {
	ts, err := foxytest.Read("testdata/k8s_contexts")
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, os.Setenv("KUBECONFIG", "./testdata/k8s_contexts/kubeconfig"))
	defer func() { require.NoError(t, os.Unsetenv("KUBECONFIG")) }()
	ts.WithExecutable("go", []string{"run", "main.go"})
	cntrl := foxytest.NewTestRunner(t)
	ts.Run(cntrl)
	ts.AssertNoErrors(cntrl)
}

func TestWithAllowedContexts(t *testing.T) {
	ts, err := foxytest.Read("testdata/allowed_contexts")
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, os.Setenv("KUBECONFIG", "./testdata/k8s_contexts/kubeconfig"))
	defer func() { require.NoError(t, os.Unsetenv("KUBECONFIG")) }()
	ts.WithExecutable("go", []string{"run", "main.go", "--allowed-contexts=allowed-ctx"})
	cntrl := foxytest.NewTestRunner(t)
	ts.Run(cntrl)
	ts.AssertNoErrors(cntrl)
}

func TestInitialize(t *testing.T) {
	ts, err := foxytest.Read("testdata/initialize")
	if err != nil {
		t.Fatal(err)
	}
	ts.WithLogging()
	ts.WithExecutable("go", []string{"run", "main.go"})
	cntrl := foxytest.NewTestRunner(t)
	ts.Run(cntrl)
	ts.AssertNoErrors(cntrl)
}

func TestLists(t *testing.T) {
	ts, err := foxytest.Read("testdata")
	if err != nil {
		t.Fatal(err)
	}
	ts.WithLogging()
	ts.WithExecutable("go", []string{"run", "main.go"})
	cntrl := foxytest.NewTestRunner(t)
	ts.Run(cntrl)
	ts.AssertNoErrors(cntrl)
}

func TestReadOnlyLists(t *testing.T) {
	ts, err := foxytest.Read("testdata/readonly")
	if err != nil {
		t.Fatal(err)
	}
	ts.WithLogging()
	ts.WithExecutable("go", []string{"run", "main.go", "--readonly"})
	cntrl := foxytest.NewTestRunner(t)
	ts.Run(cntrl)
	ts.AssertNoErrors(cntrl)
}

const k3dClusterName = "mcp-k8s-integration-test"

func TestInK3dCluster(t *testing.T) {
	testSuites := []string{
		"testdata/with_k3d",
		"internal/k8s/apps/v1/deployment",
		"internal/k8s/core/v1/pod",
		"internal/k8s/core/v1/node",
		"internal/k8s/core/v1/service",
	}

	withK3dCluster(t, k3dClusterName, func() {
		nginxImage := "nginx:1.27.3"
		busyboxImage := "busybox:1.37.0"

		preloadImage(t, nginxImage, k3dClusterName)
		preloadImage(t, busyboxImage, k3dClusterName)
		createTestNamespace(t, "test")
		createPod(t, "nginx", nginxImage)
		createPod(t, "busybox", busyboxImage, "--", "sh", "-c", "echo HELLO ; tail -f /dev/null")
		createPodService(t, "nginx", "nginx-headless", "None")

		// wait to make sure that more than a second passes for log test
		// (see more in get_k8s_pod_logs_test.yaml)
		cmd := exec.Command("kubectl", "wait", "--for=condition=Ready", "--timeout=5m", "pod", "busybox", "-n", "test")
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Second)

		for _, suite := range testSuites {
			ts, err := foxytest.Read(suite)
			if err != nil {
				t.Fatal(err)
			}

			manifestsFolder := fmt.Sprintf("%s/test_manifests", suite)

			// if exists, apply manifests specific to particular testsuite
			if _, err := os.Stat(manifestsFolder); err == nil {
				namespaceName := fmt.Sprintf("test-%s", path.Base(suite))
				createTestNamespace(t, namespaceName)
				cmd := exec.Command("kubectl", "apply", "-f", manifestsFolder)
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					t.Fatal(err)
				}
			}

			ts.WithLogging()
			ts.WithExecutable("go", []string{"run", "main.go"})
			cntrl := foxytest.NewTestRunner(t)
			ts.Run(cntrl)
			ts.AssertNoErrors(cntrl)
		}
	})
}

const kubeconfigPath = "testdata/with_k3d/kubeconfig"

func withK3dCluster(t *testing.T, name string, fn func()) {
	t.Helper()
	cmd := exec.Command("k3d", "cluster", "delete", name)
	cmd.Stderr = os.Stderr
	err := cmd.Run() // precleanup if cluster has leaked from previous test
	if err != nil {
		t.Logf("error in preclean: %v", err)
	}

	defer deleteK3dCluster(t, name)

	t.Log("creating k3d cluster", name)
	createK3dCluster(t, name)
	saveKubeconfig(t, name)
	t.Log("waiting till k3d cluster is ready")
	waitForClusterReady(t)
	fn()
}

func createTestNamespace(t *testing.T, name string) {
	t.Helper()
	cmd := exec.Command("kubectl", "create", "namespace", name)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func createPod(t *testing.T, name string, image string, args ...string) {
	t.Helper()
	allargs := []string{"-n", "test", "run", name, "--image=" + image, "--restart=Never"}
	allargs = append(allargs, args...)
	cmd := exec.Command("kubectl", allargs...)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("waiting for pod %s to be running", name)

	// wait for pod to be running
	cmd = exec.Command("kubectl", "-n", "test", "wait", "--for=condition=Ready", "--timeout=5m", "pod", name)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func createPodService(t *testing.T, podName string, serviceName string, clusterIp string) {
	t.Helper()
	cmd := exec.Command("kubectl", "expose", "-n", "test", "pod", podName, "--port", "80", "--target-port", "80", "--name", serviceName, "--cluster-ip", clusterIp)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func createK3dCluster(t *testing.T, name string) {
	t.Helper()
	cmd := exec.Command("k3d", "cluster", "create", name, "--wait", "--no-lb", "--timeout", "5m")
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	saveKubeconfig(t, name)
}

func saveKubeconfig(t *testing.T, name string) {
	require.NoError(t, os.Setenv("KUBECONFIG", kubeconfigPath))

	// write kubeconfig to file
	data, err := exec.Command("k3d", "kubeconfig", "get", name).Output()
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(kubeconfigPath, data, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func waitForClusterReady(t *testing.T) {
	// wait till all kube-system pods are running
	clients, err := k8s.GetKubeClientset()
	if err != nil {
		t.Fatal(err)
	}

	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	t.Log("waiting for kube-system pods to be created")
waiting:
	for {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for kube-system pods to be created")
		case <-ticker.C:
			pods, err := clients.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{})
			if err == nil {
				if len(pods.Items) > 0 {
					break waiting
				}
			} else {
				t.Logf("error listing pods: %v", err)
			}
		}
	}

	// This is temporarily disabled, as it makes tests slower, while we actually don't need it at the moment
	// t.Log("waiting for kube-system pods to start")
	// cmd := exec.Command("kubectl", "wait", "--for=condition=Ready", "--timeout=5m", "pod", "--all", "-n", "kube-system")
	// cmd.Stderr = os.Stderr
	// err = cmd.Run()
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

func deleteK3dCluster(t *testing.T, name string) {
	t.Helper()
	t.Log("deleting k3d cluster", name)
	cmd := exec.Command("k3d", "cluster", "delete", name)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Error(err)
	}

	t.Log("removing kubeconfig file")
	// remove kubeconfig file
	err = os.Remove(kubeconfigPath)
	if err != nil {
		t.Error(err)
	}
}

// preloadImage pulls the image and imports it into the k3d cluster
// this is needed to speed up the tests, as repeated runs would reuse
// the image from the local docker cache
func preloadImage(t *testing.T, image string, clusterName string) {
	t.Helper()
	t.Log("preloading image", image)
	cmd := exec.Command("docker", "pull", image)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("k3d", "image", "import", image, "-c", clusterName)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
