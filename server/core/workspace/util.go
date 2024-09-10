package workspace

import (
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

// LabelWorkspaceOwner sets the "workspaces.konflux.dev/is-owner" label on a
// workspace.  If workspace is nil, do nothing.
func LabelWorkspaceOwner(workspace *workspacesv1alpha1.Workspace, owner string) {
	// do nothing on an empty workspace
	if workspace == nil {
		return
	}

	var value string

	if workspace.Namespace == owner {
		value = "true"
	} else {
		value = "false"
	}

	if workspace.Labels == nil {
		workspace.Labels = map[string]string{
			workspacesv1alpha1.LabelIsOwner: value,
		}
	} else {
		workspace.Labels[workspacesv1alpha1.LabelIsOwner] = value
	}
}
