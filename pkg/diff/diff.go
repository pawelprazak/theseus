package diff

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"

	"github.com/heptio/theseus/pkg/source"
)

type Report struct {
	LeftOnly  source.ResourceSet `json:"leftOnly"`
	RightOnly source.ResourceSet `json:"rightOnly"`
	Both      source.ResourceSet `json:"both"`
}

type Options struct {
	Left          source.ResourceLister
	Right         source.ResourceLister
	Scopes        source.Includes
	LabelSelector *metav1.LabelSelector
}

func Generate(options *Options) (*Report, error) {
	leftResources, errs := options.Left.ListResources(options.Scopes, options.LabelSelector)
	if errs != nil {
		return nil, errors.NewAggregate(errs)
	}

	rightResources, errs := options.Right.ListResources(options.Scopes, options.LabelSelector)
	if errs != nil {
		return nil, errors.NewAggregate(errs)
	}

	report := &Report{
		LeftOnly:  leftResources.Except(rightResources),
		RightOnly: rightResources.Except(leftResources),
		Both:      leftResources.Intersect(rightResources),
	}

	return report, nil
}
