package k8s

import (
	"fmt"

	"github.com/strowk/mcp-k8s-go/internal/k8s/list_mapping"
)

func (p *pool) GetListMapping(k8sCtx, kind, group, version string) list_mapping.ListMapping {
	p.getInformerMutex.Lock()
	defer p.getInformerMutex.Unlock()
	key := fmt.Sprintf("%s/%s/%s/%s", k8sCtx, kind, group, version)
	res, ok := p.keyToResource[key]
	if ok {
		if res.listMapping == nil {
			mapping := findListMapping(p, res)
			if mapping != nil {
				// mapping for the same resource is not expected to change
				// , so we can cache it here to avoid finding it again later
				res.listMapping = mapping
				return mapping
			} else {
				return nil
			}
		}

	}
	return nil
}

func findListMapping(p *pool, res *resolvedResource) list_mapping.ListMapping {
	for _, resolver := range p.listMappingResolvers {
		mapping := resolver.GetListMapping(res.gvk)
		if mapping != nil {
			return mapping
		}
	}
	return nil
}
