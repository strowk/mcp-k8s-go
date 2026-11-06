package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/config"
	mock_k8s "github.com/strowk/mcp-k8s-go/internal/k8s/mock"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

type fakeGenericInformer struct {
	informer cache.SharedIndexInformer
}

func (f *fakeGenericInformer) Informer() cache.SharedIndexInformer { return f.informer }

func (f *fakeGenericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.informer.GetIndexer(), schema.GroupResource{Resource: "secrets"})
}

func newSecretInformer(t *testing.T) *fakeGenericInformer {
	t.Helper()
	informer := cache.NewSharedIndexInformer(&cache.ListWatch{}, &unstructured.Unstructured{}, 0, cache.Indexers{})
	secret := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]any{
			"name":      "s1",
			"namespace": "default",
			"annotations": map[string]any{
				"kubectl.kubernetes.io/last-applied-configuration": `{"stringData":{"password":"secret"}}`,
			},
		},
		"data": map[string]any{"password": "c2VjcmV0"},
	}}
	require.NoError(t, informer.GetIndexer().Add(secret))
	return &fakeGenericInformer{informer: informer}
}

func TestGetResourceMasksSecretsRegardlessOfSpelling(t *testing.T) {
	previous := config.GlobalOptions.MaskSecrets
	config.GlobalOptions.MaskSecrets = true
	t.Cleanup(func() { config.GlobalOptions.MaskSecrets = previous })

	for _, tc := range []struct {
		name    string
		kind    string
		version string
	}{
		{name: "lower case version", kind: "Secret", version: "v1"},
		{name: "no version", kind: "Secret", version: ""},
		{name: "upper case version", kind: "Secret", version: "V1"},
		{name: "upper case kind", kind: "SECRET", version: "V1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			poolMock := mock_k8s.NewMockClientPool(ctrl)
			poolMock.EXPECT().GetInformer("ctx", tc.kind, "", tc.version).Return(newSecretInformer(t), nil)

			tool := NewGetResourceTool(poolMock)
			resp := tool.Callback(context.Background(), map[string]any{
				"context":   "ctx",
				"namespace": "default",
				"kind":      tc.kind,
				"version":   tc.version,
				"name":      "s1",
			})

			require.NotNil(t, resp.IsError)
			require.False(t, *resp.IsError)
			require.Len(t, resp.Content, 1)
			text := resp.Content[0].(mcp.TextContent).Text
			assert.Contains(t, text, `"password":"MASKED"`)
			assert.NotContains(t, text, "c2VjcmV0")
			assert.NotContains(t, text, "last-applied-configuration")
		})
	}
}
