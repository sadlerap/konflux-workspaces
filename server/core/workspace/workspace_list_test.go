package workspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("WorkspaceList", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		lister  *MockWorkspaceLister
		request workspace.ListWorkspaceQuery
		handler workspace.ListWorkspaceHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		lister = NewMockWorkspaceLister(ctrl)
		request = workspace.ListWorkspaceQuery{}
		handler = *workspace.NewListWorkspaceHandler(lister)
	})

	AfterEach(func() { ctrl.Finish() })

	It("should not allow unauthenticated requests", func() {
		// don't set the "user" value within ctx

		response, err := handler.Handle(ctx, request)
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(fmt.Errorf("unauthenticated request")))
		Expect(response).To(BeNil())
	})

	Describe("authenticated requests", func() {
		username := "foo"
		BeforeEach(func() {
			ctx = context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
		})

		It("should allow authenticated requests", func() {
			// given
			lister.EXPECT().
				ListUserWorkspaces(ctx, username, &restworkspacesv1alpha1.WorkspaceList{}, gomock.Any()).
				Return(nil)

			// when
			response, err := handler.Handle(ctx, request)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(&workspace.ListWorkspaceResponse{
				Workspaces: restworkspacesv1alpha1.WorkspaceList{
					TypeMeta: metav1.TypeMeta{},
				},
			}))
		})

		It("should preserve owner labels", func() {
			// given
			secondUsername := "bar"
			lister.EXPECT().
				ListUserWorkspaces(ctx, username, &restworkspacesv1alpha1.WorkspaceList{}, gomock.Any()).
				Do(func(_ context.Context, username string, workspaces *restworkspacesv1alpha1.WorkspaceList, opts *client.ListOptions) {
					workspaces.Items = []restworkspacesv1alpha1.Workspace{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "default",
								Namespace: username,
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "default",
								Namespace: secondUsername,
							},
						},
					}
				}).
				Return(nil)

			// when
			response, err := handler.Handle(ctx, request)
			transform := func(ws restworkspacesv1alpha1.Workspace) metav1.ObjectMeta {
				return ws.ObjectMeta
			}

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Workspaces.Items).To(ConsistOf(
				WithTransform(transform, And(
					HaveExistingField("Name"),
					HaveExistingField("Namespace"),
					HaveExistingField("Labels"),
					HaveField("Name", "default"),
					HaveField("Namespace", username),
					HaveField("Labels", HaveKeyWithValue(restworkspacesv1alpha1.LabelIsOwner, "true")),
				)),
				WithTransform(transform, And(
					HaveExistingField("Name"),
					HaveExistingField("Namespace"),
					HaveExistingField("Labels"),
					HaveField("Name", "default"),
					HaveField("Namespace", secondUsername),
					HaveField("Labels", HaveKeyWithValue(restworkspacesv1alpha1.LabelIsOwner, "false")),
				)),
			))
		})

		It("should forward errors from the workspace reader", func() {
			// given
			error := fmt.Errorf("Failed to create workspace!")
			lister.EXPECT().
				ListUserWorkspaces(ctx, username, &restworkspacesv1alpha1.WorkspaceList{}, gomock.Any()).
				Return(error)

			// when
			response, err := handler.Handle(ctx, request)

			// then
			Expect(response).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(error))
		})
	})
})
