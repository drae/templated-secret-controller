// Based on work by Carvel - that work Copyright 2024 The Carvel Authors.
// Re-organized by starstreak.dev - that work Copyright 2025 starstreak.dev
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	templatedsecretv1alpha1 "github.com/drae/templated-secret-controller/pkg/apis/templatedsecret/v1alpha1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// SecretTemplateLister helps list SecretTemplates.
// All objects returned here must be treated as read-only.
type SecretTemplateLister interface {
	// List lists all SecretTemplates in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*templatedsecretv1alpha1.SecretTemplate, err error)
	// SecretTemplates returns an object that can list and get SecretTemplates.
	SecretTemplates(namespace string) SecretTemplateNamespaceLister
	SecretTemplateListerExpansion
}

// secretTemplateLister implements the SecretTemplateLister interface.
type secretTemplateLister struct {
	listers.ResourceIndexer[*templatedsecretv1alpha1.SecretTemplate]
}

// NewSecretTemplateLister returns a new SecretTemplateLister.
func NewSecretTemplateLister(indexer cache.Indexer) SecretTemplateLister {
	return &secretTemplateLister{listers.New[*templatedsecretv1alpha1.SecretTemplate](indexer, templatedsecretv1alpha1.Resource("secrettemplate"))}
}

// SecretTemplates returns an object that can list and get SecretTemplates.
func (s *secretTemplateLister) SecretTemplates(namespace string) SecretTemplateNamespaceLister {
	return secretTemplateNamespaceLister{listers.NewNamespaced[*templatedsecretv1alpha1.SecretTemplate](s.ResourceIndexer, namespace)}
}

// SecretTemplateNamespaceLister helps list and get SecretTemplates.
// All objects returned here must be treated as read-only.
type SecretTemplateNamespaceLister interface {
	// List lists all SecretTemplates in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*templatedsecretv1alpha1.SecretTemplate, err error)
	// Get retrieves the SecretTemplate from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*templatedsecretv1alpha1.SecretTemplate, error)
	SecretTemplateNamespaceListerExpansion
}

// secretTemplateNamespaceLister implements the SecretTemplateNamespaceLister
// interface.
type secretTemplateNamespaceLister struct {
	listers.ResourceIndexer[*templatedsecretv1alpha1.SecretTemplate]
}
