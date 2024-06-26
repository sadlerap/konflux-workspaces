package readclient

import (
	"context"
	"fmt"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
)

var _ workspace.WorkspaceLister = &ReadClient{}

// ListUserWorkspaces Returns all the workspaces the user has access to
func (c *ReadClient) ListUserWorkspaces(
	ctx context.Context,
	user string,
	objs *restworkspacesv1alpha1.WorkspaceList,
	opts ...client.ListOption,
) error {
	// retrieve workspaces visible to user
	iww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := c.internalClient.ListAsUser(ctx, user, &iww); err != nil {
		return kerrors.NewInternalError(fmt.Errorf("error retrieving the list of workspaces for user %v", user))
	}

	// map list options to InternalWorkspaces
	listOpts, err := mapListOptions(opts...)
	if err != nil {
		return err
	}

	// filter internal workspaces
	fiww := filterByLabels(&iww, listOpts)

	// map back to Workspaces
	ww, err := c.mapper.InternalWorkspaceListToWorkspaceList(fiww)
	if err != nil {
		return kerrors.NewInternalError(fmt.Errorf("error retrieving the list of workspaces for user %v", user))
	}

	ww.DeepCopyInto(objs)
	return nil
}

func filterByLabels(ww *workspacesv1alpha1.InternalWorkspaceList, listOpts *client.ListOptions) *workspacesv1alpha1.InternalWorkspaceList {
	rww := workspacesv1alpha1.InternalWorkspaceList{}
	for _, w := range ww.Items {
		// selection
		if !matchesListOpts(listOpts, w.GetLabels()) {
			continue
		}

		rww.Items = append(rww.Items, w)
	}
	return &rww
}

func mapListOptions(opts ...client.ListOption) (*client.ListOptions, error) {
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	if listOpts.LabelSelector == nil {
		return nil, nil
	}

	rr, _ := listOpts.LabelSelector.Requirements()
	for _, ls := range rr {
		if strings.HasPrefix(ls.Key(), workspacesv1alpha1.LabelInternalDomain) {
			return nil, fmt.Errorf("invalid label selector: key '%s' is reserved", ls.Key())
		}
	}

	return &listOpts, nil
}

func matchesListOpts(
	listOpts *client.ListOptions,
	objLabels map[string]string,
) bool {
	return objLabels == nil || listOpts == nil || listOpts.LabelSelector == nil ||
		listOpts.LabelSelector.Matches(labels.Set(objLabels))
}
