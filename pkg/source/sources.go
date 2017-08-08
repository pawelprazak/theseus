package source

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceLister interface {
	ListResources(scopes Includes, labelSelector *metav1.LabelSelector) (ResourceSet, []error)
}

func Get(sourceType, location string) (ResourceLister, error) {
	switch sourceType {
	case "cluster":
		return Cluster(location)
	case "repo":
		return Repo(location)
	case "backup":
		return Backup(location)
	case "directory":
		return Directory(location)
	default:
		return nil, errors.New("source type must be one of cluster, repo, backup, directory")
	}
}

func getScopeKey(scope string) string {
	if scope == "" {
		return "cluster"
	}

	return fmt.Sprintf("ns:%s", scope)
}
