package trilio

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctlr "sigs.k8s.io/controller-runtime/pkg/client"
)

type TargetAnalyzer struct {
	TargetName string
}

func (c *TargetAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	ctx := a.Context
	client := a.Client.GetCtrlClient()
	namespace := a.Namespace
	kind := TargetValidate
	var failures []common.Failure
	gvk := schema.GroupVersionKind{
		Group:   "triliovault.trilio.io",
		Version: "v1",
		Kind:    "Target",
	}
	opts := &ctlr.ListOptions{
		Namespace: namespace,
	}
	unstructuredList := &unstructured.UnstructuredList{}
	unstructuredList.SetAPIVersion(gvk.GroupVersion().String())
	unstructuredList.SetKind(gvk.Kind)

	err := client.List(ctx, unstructuredList, opts)
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}
	for _, item := range unstructuredList.Items {
		// Access status.status field using accessors
		status, found, err := unstructured.NestedString(item.Object, "status", "status")
		if err != nil {
			color.Red("Error accessing status for %s: %v\n", item.GetName(), err)
			continue
		}
		if !found {
			color.Red("Resource %s has no status.status field\n", item.GetName())
			continue
		}

		if status != "Available" {
			conditions, found, err := unstructured.NestedSlice(item.Object, "status", "condition")
			if err != nil {
				color.Red("Error accessing status for %s: %v\n", item.GetName(), err)
				continue
			}
			if !found || len(conditions) == 0 {
				color.Red("Resource %s has no status.condition field\n", item.GetName())
				continue
			}
			reason, found, err := unstructured.NestedString(conditions[0].(map[string]interface{}), "reason")
			if err != nil {
				color.Red("Error accessing status for %s: %v\n", item.GetName(), err)
				continue
			}
			if !found {
				color.Red("Resource %s has no condition.reason field\n", item.GetName())
				continue
			}

			failures = append(failures, common.Failure{
				Text: reason,
			})
			preAnalysis[fmt.Sprintf("%s/%s", item.GetName(), item.GetNamespace())] = common.PreAnalysis{
				TargetAnalyzer: common.TargetAnalyzer{
					TargetName: item.GetName(),
				},
				FailureDetails: failures,
			}
		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, _ := util.GetParent(a.Client, value.Pod.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
