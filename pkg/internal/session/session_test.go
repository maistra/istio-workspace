package session_test

import (
	"time"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	testclient "github.com/maistra/istio-workspace/pkg/client/clientset/versioned/fake"
	"github.com/maistra/istio-workspace/pkg/internal/session"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Session operations", func() {

	const (
		telepresence  = "telepresence"
		preparedImage = "prepared-image"
	)
	Context("handlers", func() {
		var (
			objects       []runtime.Object
			client        *session.Client
			opts          session.Options
			updateRemover func()
		)
		BeforeEach(func() {
			duration := 10 * time.Millisecond
			opts = session.Options{
				NamespaceName:  "test-namespace",
				SessionName:    "test-session",
				DeploymentName: "test-deployment",
				Strategy:       preparedImage,
				Duration:       &duration,
			}
		})
		JustBeforeEach(func() {
			client, _ = session.NewClient(testclient.NewSimpleClientset(objects...), "test-namespace")
			updateRemover = addSessionRefStatus(client, opts.SessionName)
		})
		AfterEach(func() {
			updateRemover()
		})
		Context("create", func() {

			It("should create a new session if non found", func() {
				// given - no exiting sessions
				// when - adding a ref to a session
				_, remove, err := session.CreateOrJoinHandler(opts, client)
				defer remove()
				Expect(err).ToNot(HaveOccurred())

				// then - a session should exist
				sess, err := client.Get(opts.SessionName)
				Expect(err).ToNot(HaveOccurred())

				Expect(sess.Spec.Refs).To(HaveLen(1))
			})
		})
		Context("join", func() {
			BeforeEach(func() {
				objects = []runtime.Object{
					&istiov1alpha1.Session{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "maistra.io/v1alpha1",
							Kind:       "Session",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      opts.SessionName,
							Namespace: opts.NamespaceName,
						},
						Spec: istiov1alpha1.SessionSpec{
							Refs: []istiov1alpha1.Ref{
								{Name: opts.DeploymentName + "-1", Strategy: opts.Strategy, Args: opts.StrategyArgs},
							},
						},
					}}
			})

			It("should join a session if existing name found", func() {
				// given - an existing session
				// when - adding a ref to a session with the same name
				_, remove, err := session.CreateOrJoinHandler(opts, client)
				defer remove()
				Expect(err).ToNot(HaveOccurred())

				// then - expect there to be two refs
				sess, err := client.Get(opts.SessionName)
				Expect(err).ToNot(HaveOccurred())

				Expect(sess.Spec.Refs).To(HaveLen(2))
			})

			It("should revert ref to previous state on remove", func() {
				// given - an existing ref of prepared-image

				// when - update the existing ref with telepresence
				opts.Revert = true
				opts.DeploymentName = opts.DeploymentName + "-1"
				opts.Strategy = telepresence

				_, remove, err := session.CreateOrJoinHandler(opts, client)
				Expect(err).ToNot(HaveOccurred())

				// then - expect the strategy to be updated
				sess, err := client.Get(opts.SessionName)
				Expect(err).ToNot(HaveOccurred())

				Expect(sess.Spec.Refs).To(HaveLen(1))
				Expect(sess.Spec.Refs[0].Strategy).To(Equal(telepresence))

				// when - the ref take over is removed
				remove()

				// then - expect the ref to be back to prepared-image
				sess, err = client.Get(opts.SessionName)
				Expect(err).ToNot(HaveOccurred())

				Expect(sess.Spec.Refs).To(HaveLen(1))
				Expect(sess.Spec.Refs[0].Strategy).To(Equal(preparedImage))
			})

			It("should not revert if ref was never updated", func() {
				// given - an existing ref of prepared-image

				// when - update the existing ref with telepresence
				opts.Revert = true
				opts.Strategy = telepresence

				_, remove, err := session.CreateOrJoinHandler(opts, client)
				Expect(err).ToNot(HaveOccurred())

				// then - expect another ref to have been added
				sess, err := client.Get(opts.SessionName)
				Expect(err).ToNot(HaveOccurred())

				Expect(sess.Spec.Refs).To(HaveLen(2))

				// when - the ref is removed
				remove()

				// then - expect the only ref to be prepared-image
				sess, err = client.Get(opts.SessionName)
				Expect(err).ToNot(HaveOccurred())

				Expect(sess.Spec.Refs).To(HaveLen(1))
				Expect(sess.Spec.Refs[0].Strategy).To(Equal(preparedImage))
			})
		})
		Context("remove", func() {
			BeforeEach(func() {
				objects = []runtime.Object{
					&istiov1alpha1.Session{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "maistra.io/v1alpha1",
							Kind:       "Session",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      opts.SessionName,
							Namespace: opts.NamespaceName,
						},
						Spec: istiov1alpha1.SessionSpec{
							Refs: []istiov1alpha1.Ref{
								{Name: opts.DeploymentName + "-1", Strategy: opts.Strategy, Args: opts.StrategyArgs},
							},
						},
					}}
			})

			It("should join a session if existing name found", func() {
				// given - an existing session
				opts.DeploymentName = opts.DeploymentName + "-1"

				// when - removing a ref from a session
				_, remove, err := session.RemoveHandler(opts, client)
				Expect(err).ToNot(HaveOccurred())

				remove()

				// then - expect there to be no session
				_, err = client.Get(opts.SessionName)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("route parsing", func() {

		It("should return nil with no error on empty string", func() {
			r, err := session.ParseRoute("")
			Expect(err).ToNot(HaveOccurred())
			Expect(r).To(BeNil())
		})

		It("should error on wrong type format", func() {
			_, err := session.ParseRoute("a=b")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("route in wrong format"))
		})

		It("should error on wrong value format", func() {
			_, err := session.ParseRoute("header:a-b")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("route in wrong format"))
		})

		It("should return a valid route", func() {
			r, err := session.ParseRoute("header:a=b")
			Expect(err).ToNot(HaveOccurred())
			Expect(r).ToNot(BeNil())

			Expect(r.Type).To(Equal("header"))
			Expect(r.Name).To(Equal("a"))
			Expect(r.Value).To(Equal("b"))
		})
	})
})

// helper function to mimic the server side reacting to the session object.
func addSessionRefStatus(c *session.Client, sessionName string) func() {
	done := false
	go func() {
		for {
			if done {
				break
			}
			time.Sleep(5 * time.Millisecond)
			sess, err := c.Get(sessionName)
			if err != nil {
				continue
			}
			sess.Status.Route = &istiov1alpha1.Route{
				Type:  "header",
				Name:  "x-workspace-route",
				Value: "xxxx",
			}
			for _, ref := range sess.Spec.Refs {
				found := false
				for _, status := range sess.Status.Refs {
					if status.Name == ref.Name {
						found = true
					}
				}
				if found {
					continue
				}
				kind := "Deployment"
				name := "test-deployment-clone"
				action := "created"
				sess.Status.Refs = append(sess.Status.Refs, &istiov1alpha1.RefStatus{
					Ref: istiov1alpha1.Ref{
						Name: ref.Name,
					},
					Resources: []*istiov1alpha1.RefResource{
						{
							Kind:   &kind,
							Name:   &name,
							Action: &action,
						},
					},
				})
			}
			updateErr := c.Update(sess)
			Expect(updateErr).ToNot(HaveOccurred())
		}
	}()
	return func() {
		done = true
	}
}
