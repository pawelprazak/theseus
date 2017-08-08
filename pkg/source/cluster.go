package source

import (
	"fmt"
	"os"

	"github.com/heptio/ark/pkg/util/collections"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type cluster struct {
	kubeClient      *kubernetes.Clientset
	discoveryClient discovery.DiscoveryInterface
	clientPool      dynamic.ClientPool
}

var _ ResourceLister = &cluster{}

// Cluster connects to a running Kubernetes cluster with the provided
// configuration and returns a Source for the cluster.
func Cluster(kubeconfig string) (ResourceLister, error) {
	config, err := getConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &cluster{
		kubeClient:      kubeClient,
		discoveryClient: kubeClient.Discovery(),
		clientPool:      dynamic.NewDynamicClientPool(config),
	}, nil
}

func (c *cluster) ListResources(scopes Includes, labelSelector *metav1.LabelSelector) (ResourceSet, []error) {
	groupVersions, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, []error{err}
	}

	groupVersions = discovery.FilteredBy(discovery.SupportsAllVerbs{Verbs: []string{"list"}}, groupVersions)

	var (
		errs []error
		res  = NewResourceSet()
	)

	nses, err := c.kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, []error{err}
	}

	// FormatLabelSelector returns "<none>" for a nil/empty LabelSelector
	// which is not what we want.
	labelSelectorString := ""
	if labelSelector != nil {
		labelSelectorString = metav1.FormatLabelSelector(labelSelector)
	}

	for _, groupVersion := range groupVersions {
		for _, resource := range groupVersion.APIResources {
			gvk, err := groupVersionKind(groupVersion, resource)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			client, err := c.clientPool.ClientForGroupVersionKind(gvk)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			if resource.Namespaced {
				for _, ns := range nses.Items {
					if !scopes.ShouldInclude(ns.Name) {
						continue
					}

					resourceErrs := processResourceForScope(&resource, gvk, ns.Name, client, labelSelectorString, res)
					errs = append(errs, resourceErrs...)
				}
			} else {
				if !scopes.ShouldInclude("cluster") {
					continue
				}

				resourceErrs := processResourceForScope(&resource, gvk, "", client, labelSelectorString, res)
				errs = append(errs, resourceErrs...)
			}
		}
	}

	return res, errs
}

func processResourceForScope(
	resource *metav1.APIResource,
	groupVersionKind schema.GroupVersionKind,
	scope string,
	client dynamic.Interface,
	labelSelector string,
	resourceSet ResourceSet,
) []error {
	var errs []error

	resources, err := client.Resource(resource, scope).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	items, err := meta.ExtractList(resources)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	if addErrs := addItems(resourceSet, scope, groupVersionKind, items); addErrs != nil {
		errs = append(errs, addErrs...)
	}

	return errs
}

func addItems(resourceSet ResourceSet, namespace string, groupVersionKind schema.GroupVersionKind, items []runtime.Object) []error {
	var (
		errs     []error
		scopeKey = getScopeKey(namespace)
	)

	for _, item := range items {
		unstructured, ok := item.(runtime.Unstructured)
		if !ok {
			errs = append(errs, fmt.Errorf("unexpected type %T", item))
			continue
		}

		name, err := collections.GetString(unstructured.UnstructuredContent(), "metadata.name")
		if err != nil {
			errs = append(errs, err)
			continue
		}

		resourceSet.Add(scopeKey, groupVersionKind.String(), name, unstructured.UnstructuredContent())
	}

	return errs
}

func groupVersionKind(groupVersion *metav1.APIResourceList, resource metav1.APIResource) (schema.GroupVersionKind, error) {
	gv, err := schema.ParseGroupVersion(groupVersion.GroupVersion)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	return gv.WithKind(resource.Kind), nil
}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
