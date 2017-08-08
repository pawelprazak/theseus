package source

import (
	"fmt"
	"strings"

	"github.com/yudai/gojsondiff"
)

const keySeparator string = "||"

// ResourceKey is a string used as a unique key in a map
// for Kubernetes resources. Contains scope, GVK, and name
// separated by a delimiter.
type ResourceKey string

// Parts returns the constituent parts of a ResourceKey (scope,
// GVK, name), or an error if the ResourceKey is not valid.
func (rk ResourceKey) Parts() (string, string, string, error) {
	return splitResourceKey(rk)
}

func getResourceKey(scope, gvk, name string) ResourceKey {
	return ResourceKey(strings.Join([]string{scope, gvk, name}, keySeparator))
}

func splitResourceKey(key ResourceKey) (string, string, string, error) {
	parts := strings.Split(string(key), keySeparator)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("key %s is not formatted properly", key)
	}

	return parts[0], parts[1], parts[2], nil
}

// ObjectDiff contains information about a Kubernetes object
// in an unstructured format, along with diff information from
// comparing this object to another.
type ObjectDiff struct {
	Object map[string]interface{} `json:"object"`
	Diff   gojsondiff.Diff        `json:"-"`
}

// ResourceSet is a mapping from ResourceKeys to ObjectDiffs for
// Kubernetes objects.
type ResourceSet map[ResourceKey]*ObjectDiff

// NewResourceSet creates a new empty ResourceSet.
func NewResourceSet() ResourceSet {
	return make(ResourceSet)
}

// Add inserts information about a given Kubernetes object into the ResourceSet.
func (rs ResourceSet) Add(scope, gvk, name string, obj map[string]interface{}) ResourceSet {
	key := getResourceKey(scope, gvk, name)
	rs[key] = &ObjectDiff{obj, nil}
	return rs
}

// Get returns information about a specified Kubernetes object, if it exists.
func (rs ResourceSet) Get(scope, gvk, name string) *ObjectDiff {
	key := getResourceKey(scope, gvk, name)
	return rs[key]
}

// Except returns a new ResourceSet containing all items that are
// in the left ResourceSet but not the right (based on ResourceKey).
func (left ResourceSet) Except(right ResourceSet) ResourceSet {
	except := NewResourceSet()

	for k := range left {
		if right[k] == nil {
			except[k] = left[k]
		}
	}

	return except
}

// Intersect returns a new ResourceSet containing all items that are
// in both left and right, including linewise diff data (stored in the
// Diff field of the ObjectDiff values).
func (left ResourceSet) Intersect(right ResourceSet) ResourceSet {
	intersect := NewResourceSet()

	for key := range left {
		if _, found := right[key]; found {
			intersect[key] = &ObjectDiff{
				left[key].Object,
				gojsondiff.New().CompareObjects(left[key].Object, right[key].Object),
			}
		}
	}

	return intersect
}
