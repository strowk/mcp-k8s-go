package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SanitizeObjectMeta(object *metav1.ObjectMeta) {
	// exclude managed fields, since they are not relevant for users and would only
	// clutter the context window with irrelevant information
	object.ManagedFields = nil
}
