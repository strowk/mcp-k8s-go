package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	mock_k8s "github.com/strowk/mcp-k8s-go/internal/k8s/mock"
	"github.com/strowk/mcp-k8s-go/internal/tests"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodLogs(t *testing.T) {
	cntr := gomock.NewController(t)
	poolMock := mock_k8s.NewMockClientPool(cntr)

	tool := NewPodLogsTool(poolMock)
	t.Run("Call with invalid type of previousContainer", func(t *testing.T) {
		args := map[string]any{
			"context":           "context",
			"namespace":         "namespace",
			"pod":               "pod",
			"previousContainer": "invalid",
		}
		resp := tool.Callback(context.Background(), args)
		if assert.NotNil(t, resp.IsError) {
			assert.True(t, *resp.IsError)
		}
	})

	t.Run("Call with boolean within string for previousContainer", func(t *testing.T) {
		args := map[string]any{
			"context":           "context",
			"namespace":         "namespace",
			"pod":               "pod",
			"previousContainer": "true", // this is what Inspector gives us, this might be a bug
		}
		poolMock.EXPECT().GetClientset("context").Return(fake.NewClientset(
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "namespace",
				},
			},
		), nil)
		resp := tool.Callback(context.Background(), args)
		tests.AssertTextContentContainsInFirstString(t, "fake logs", resp.Content)
	})

	t.Run("Call with empty string for previousContainer", func(t *testing.T) {
		args := map[string]any{
			"context":           "context",
			"namespace":         "namespace",
			"pod":               "pod",
			"previousContainer": "",
		}
		poolMock.EXPECT().GetClientset("context").Return(fake.NewClientset(
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "namespace",
				},
			},
		), nil)
		resp := tool.Callback(context.Background(), args)
		tests.AssertTextContentContainsInFirstString(t, "fake logs", resp.Content)
	})
}
