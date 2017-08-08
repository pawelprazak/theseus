package source

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type backup struct {
	file *os.File
}

var _ ResourceLister = &backup{}

// Backup connects to an Ark backup at the provided file path
// and returns a Source for the backup.
func Backup(fileName string) (ResourceLister, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	return &backup{
		file: file,
	}, nil
}

func (b *backup) ListResources(scopes Includes, labelSelector *metav1.LabelSelector) (ResourceSet, []error) {
	gzr, err := gzip.NewReader(b.file)
	if err != nil {
		return nil, []error{err}
	}
	defer gzr.Close()

	var (
		tarReader = tar.NewReader(gzr)
		res       = NewResourceSet()
		errs      []error
	)

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

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, []error{err}
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		pathParts := strings.Split(header.Name, string(os.PathSeparator))

		var scopeKey string

		if pathParts[0] == "cluster" {
			if !scopes.ShouldInclude(pathParts[0]) {
				continue
			}

			scopeKey = getScopeKey("")
		} else if pathParts[0] == "namespaces" {
			if !scopes.ShouldInclude(pathParts[1]) {
				continue
			}

			scopeKey = getScopeKey(pathParts[1])
		}

		// read file, get group/version/kind/name
		contents, err := ioutil.ReadAll(tarReader)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		var obj unstructured.Unstructured

		if err := json.Unmarshal(contents, &obj); err != nil {
			errs = append(errs, err)
			continue
		}

		if !selector.Matches(labels.Set(obj.GetLabels())) {
			continue
		}

		gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		gvk := gv.WithKind(obj.GetKind())

		res.Add(scopeKey, gvk.String(), obj.GetName(), obj.UnstructuredContent())
	}

	return res, errs
}
