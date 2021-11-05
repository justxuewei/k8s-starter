/*
samplecrd by justxuewei
inspired by https://time.geekbang.org/column/article/41876
*/
// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	internalinterfaces "github.com/justxuewei/k8s_starter/samplecrd/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Networks returns a NetworkInformer.
	Networks() NetworkInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Networks returns a NetworkInformer.
func (v *version) Networks() NetworkInformer {
	return &networkInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
