package workspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"sigs.k8s.io/controller-runtime/pkg/client"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("WorkspaceRead", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		reader  *MockWorkspaceReader
		request workspace.ReadWorkspaceQuery
		handler workspace.ReadWorkspaceHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		reader = NewMockWorkspaceReader(ctrl)
		request = workspace.ReadWorkspaceQuery{}
		handler = *workspace.NewReadWorkspaceHandler(reader)
	})

	AfterEach(func() { ctrl.Finish() })

	It("should not allow unauthenticated requests", func() {
		// don't set the "user" value within ctx

		response, err := handler.Handle(ctx, request)
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(fmt.Errorf("unauthenticated request")))
		Expect(response).To(BeNil())
	})

	When("the request is authenticated", func() {
		// given
		username := "foo"

		BeforeEach(func() {
			ctx = context.WithValue(ctx, ccontext.UserSignupComplaintNameKey, username)
			request.Owner = username
			request.Name = "default"
		})

		It("should allow authenticated requests", func() {
			// given
			reader.EXPECT().
				ReadUserWorkspace(ctx,
					username,
					request.Owner,
					request.Name,
					&restworkspacesv1alpha1.Workspace{},
					[]client.GetOption{}).
				Return(nil)

			// when
			_, err := handler.Handle(ctx, request)

			// then
			Expect(err).NotTo(HaveOccurred())
		})

		It("should forward errors from the workspace reader", func() {
			// given
			error := fmt.Errorf("Failed to create workspace!")
			reader.EXPECT().
				ReadUserWorkspace(ctx,
					username,
					request.Owner,
					request.Name,
					&restworkspacesv1alpha1.Workspace{},
					[]client.GetOption{}).
				Return(error)

			// when
			response, err := handler.Handle(ctx, request)

			// then
			Expect(response).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(error))
		})

		When("a user requests an owned workspace", func() {
			It("should set the owned label to true", func() {
				// given
				reader.EXPECT().ReadUserWorkspace(ctx,
					username,
					request.Owner,
					request.Name,
					&restworkspacesv1alpha1.Workspace{},
					[]client.GetOption{}).
					Do(func(_ context.Context, _, owner, name string, ws *restworkspacesv1alpha1.Workspace, _ ...client.GetOption) {
						ws.SetName(name)
						ws.SetNamespace(owner)
					}).
					Return(nil)

				// when
				response, err := handler.Handle(ctx, request)

				// then
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Workspace).NotTo(BeNil())
				Expect(response.Workspace.Labels).To(HaveKeyWithValue(restworkspacesv1alpha1.LabelIsOwner, "true"))
			})

			It("should preserve existing labels", func() {
				// given
				reader.EXPECT().ReadUserWorkspace(ctx,
					username,
					request.Owner,
					request.Name,
					&restworkspacesv1alpha1.Workspace{},
					[]client.GetOption{}).
					Do(func(_ context.Context, _, owner, name string, ws *restworkspacesv1alpha1.Workspace, _ ...client.GetOption) {
						ws.SetName(name)
						ws.SetNamespace(owner)
						ws.Labels = map[string]string{"foo": "bar"}
					}).
					Return(nil)

				// when
				response, err := handler.Handle(ctx, request)

				// then
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Workspace).NotTo(BeNil())
				Expect(response.Workspace.Labels).To(HaveKeyWithValue(restworkspacesv1alpha1.LabelIsOwner, "true"))
				Expect(response.Workspace.Labels).To(HaveKeyWithValue("foo", "bar"))
			})
		})

		When("a user requests another workspace", func() {
			actualOwner := "bar"

			It("should set the owned label to false", func() {
				// given
				request.Owner = actualOwner
				reader.EXPECT().ReadUserWorkspace(ctx,
					username,
					request.Owner,
					request.Name,
					&restworkspacesv1alpha1.Workspace{},
					[]client.GetOption{}).
					Do(func(_ context.Context, _, owner, name string, ws *restworkspacesv1alpha1.Workspace, _ ...client.GetOption) {
						ws.SetName(name)
						ws.SetNamespace(owner)
					}).
					Return(nil)

				// when
				response, err := handler.Handle(ctx, request)

				// then
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Workspace).NotTo(BeNil())
				Expect(response.Workspace.Labels).To(HaveKeyWithValue(restworkspacesv1alpha1.LabelIsOwner, "false"))
			})
		})
	})
})
