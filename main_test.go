package main

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/strowk/foxy-contexts/pkg/foxytest"
	"github.com/strowk/mcp-k8s-go/internal/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListContexts(t *testing.T) {
	ts, err := foxytest.Read("testdata/k8s_contexts")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("KUBECONFIG", "./testdata/k8s_contexts/kubeconfig")
	defer os.Unsetenv("KUBECONFIG")
	ts.WithExecutable("go", []string{"run", "main.go"})
	cntrl := foxytest.NewTestRunner(t)
	ts.Run(cntrl)
	ts.AssertNoErrors(cntrl)
}

func TestListTools(t *testing.T) {
	ts, err := foxytest.Read("testdata")
	if err != nil {
		t.Fatal(err)
	}
	ts.WithExecutable("go", []string{"run", "main.go"})
	cntrl := foxytest.NewTestRunner(t)
	ts.Run(cntrl)
	ts.AssertNoErrors(cntrl)
}

const k3dClusterName = "mcp-k8s-integration-test"

func TestInK3dCluster(t *testing.T) {
	// if os.Getenv("CI") != "" {
	// 	t.Skip("Skipping k3d tests in CI for now")
	// }
	ts, err := foxytest.Read("testdata/with_k3d")
	if err != nil {
		t.Fatal(err)
	}

	withK3dCluster(t, k3dClusterName, func() {
		preloadImage(t, "nginx:latest", k3dClusterName)
		preloadImage(t, "busybox:latest", k3dClusterName)
		createPod(t, "nginx", "nginx:latest")
		createPod(t, "busybox", "busybox:latest", "--", "sh", "-c", "echo HELLO ; tail -f /dev/null")
		ts.WithLogging()
		ts.WithExecutable("go", []string{"run", "main.go"})
		cntrl := foxytest.NewTestRunner(t)
		ts.Run(cntrl)
		ts.AssertNoErrors(cntrl)
	})
}

const kubeconfigPath = "testdata/with_k3d/kubeconfig"

func withK3dCluster(t *testing.T, name string, fn func()) {
	t.Helper()
	cmd := exec.Command("k3d", "cluster", "delete", name)
	cmd.Stderr = os.Stderr
	cmd.Run() // precleanup if cluster has leaked from previous test

	defer deleteK3dCluster(t, name)

	t.Log("creating k3d cluster", name)
	createK3dCluster(t, name)
	saveKubeconfig(t, name)
	t.Log("waiting till k3d cluster is ready")
	waitForClusterReady(t)
	fn()
}

func createPod(t *testing.T, name string, image string, args ...string) {
	t.Helper()
	allargs := []string{"-n", "default", "run", name, "--image=" + image, "--restart=Never"}
	allargs = append(allargs, args...)
	cmd := exec.Command("kubectl", allargs...)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("waiting for pod %s to be running", name)

	// wait for pod to be running
	cmd = exec.Command("kubectl", "-n", "default", "wait", "--for=condition=Ready", "--timeout=5m", "pod", name)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
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
	os.Setenv("KUBECONFIG", kubeconfigPath)

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
			if err != nil {
				t.Fatal(err)
			}
			if len(pods.Items) > 0 {
				break waiting
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
