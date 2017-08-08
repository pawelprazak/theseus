package source

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-yaml/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type directory struct {
	path string
}

var _ ResourceLister = &directory{}

// Directory connects to a local directory containing Kubernetes
// resource definitions and returns a ResourceLister for the
// directory.
func Directory(path string) (ResourceLister, error) {
	if _, err := ioutil.ReadDir(path); err != nil {
		return nil, err
	}

	return &directory{
		path: path,
	}, nil
}

func (d *directory) ListResources(scopes Includes, labelSelector *metav1.LabelSelector) (ResourceSet, []error) {
	// metav1.LabelSelectorAsSelector converts a nil LabelSelector to a
	// Nothing Selector, i.e. a selector that matches nothing. We want
	// a selector that matches everything. This can be accomplished by
	// passing a non-nil empty LabelSelector.
	if labelSelector == nil {
		labelSelector = &metav1.LabelSelector{}
	}

	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, []error{err}
	}

	res := NewResourceSet()

	err = filepath.Walk(d.path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		var ext string
		if ext = filepath.Ext(path); ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var obj unstructured.Unstructured

		switch ext {
		case ".json":
			if err := json.Unmarshal(contents, &obj); err != nil {
				return err
			}
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(contents, &obj); err != nil {
				return err
			}
		}

		if !selector.Matches(labels.Set(obj.GetLabels())) {
			return nil
		}

		gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())
		if err != nil {
			return err
		}
		gvk := gv.WithKind(obj.GetKind())

		res.Add(getScopeKey(obj.GetNamespace()), gvk.String(), obj.GetName(), obj.UnstructuredContent())

		return nil
	})

	if err != nil {
		return nil, []error{err}
	}

	return res, nil
}
