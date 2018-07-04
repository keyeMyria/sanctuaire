package kubernetes

import (
	"testing"
	"time"

	"github.com/keel-hq/keel/internal/k8s"
	"github.com/keel-hq/keel/types"

	apps_v1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckRequestedApproval(t *testing.T) {
	fp := &fakeImplementer{}
	fp.namespaces = &v1.NamespaceList{
		Items: []v1.Namespace{
			v1.Namespace{
				meta_v1.TypeMeta{},
				meta_v1.ObjectMeta{Name: "xxxx"},
				v1.NamespaceSpec{},
				v1.NamespaceStatus{},
			},
		},
	}
	deployments := []*apps_v1.Deployment{
		{
			meta_v1.TypeMeta{},
			meta_v1.ObjectMeta{
				Name:        "dep-1",
				Namespace:   "xxxx",
				Labels:      map[string]string{types.KeelPolicyLabel: "all", types.KeelMinimumApprovalsLabel: "1"},
				Annotations: map[string]string{},
			},
			apps_v1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							v1.Container{
								Image: "gcr.io/v2-namespace/hello-world:1.1.1",
							},
						},
					},
				},
			},
			apps_v1.DeploymentStatus{},
		},
	}

	grs := MustParseGRS(deployments)
	grc := &k8s.GenericResourceCache{}
	grc.Add(grs...)

	approver := approver()
	provider, err := NewProvider(fp, &fakeSender{}, approver, grc)
	if err != nil {
		t.Fatalf("failed to get provider: %s", err)
	}
	// creating "new version" event
	repo := types.Repository{
		Name: "gcr.io/v2-namespace/hello-world",
		Tag:  "1.1.2",
	}

	deps, err := provider.processEvent(&types.Event{Repository: repo})
	if err != nil {
		t.Errorf("failed to get deployments: %s", err)
	}

	if len(deps) != 0 {
		t.Errorf("expected to find 0 updated deployment but found %d", len(deps))
	}

	// checking approvals
	approval, err := provider.approvalManager.Get("deployment/xxxx/dep-1:1.1.2")
	if err != nil {
		t.Fatalf("failed to find approval, err: %s", err)
	}

	if approval.Provider != types.ProviderTypeKubernetes {
		t.Errorf("wrong provider: %s", approval.Provider)
	}
}

func TestApprovedCheck(t *testing.T) {
	fp := &fakeImplementer{}
	fp.namespaces = &v1.NamespaceList{
		Items: []v1.Namespace{
			v1.Namespace{
				meta_v1.TypeMeta{},
				meta_v1.ObjectMeta{Name: "xxxx"},
				v1.NamespaceSpec{},
				v1.NamespaceStatus{},
			},
		},
	}
	deployments := []*apps_v1.Deployment{
		{
			meta_v1.TypeMeta{},
			meta_v1.ObjectMeta{
				Name:        "dep-1",
				Namespace:   "xxxx",
				Labels:      map[string]string{types.KeelPolicyLabel: "all", types.KeelMinimumApprovalsLabel: "1"},
				Annotations: map[string]string{},
			},
			apps_v1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							v1.Container{
								Image: "gcr.io/v2-namespace/hello-world:1.1.1",
							},
						},
					},
				},
			},
			apps_v1.DeploymentStatus{},
		},
	}
	grs := MustParseGRS(deployments)
	grc := &k8s.GenericResourceCache{}
	grc.Add(grs...)

	approver := approver()
	provider, err := NewProvider(fp, &fakeSender{}, approver, grc)
	if err != nil {
		t.Fatalf("failed to get provider: %s", err)
	}

	// approving event
	err = provider.approvalManager.Create(&types.Approval{
		Identifier:    "deployment/xxxx/dep-1:1.1.2",
		VotesReceived: 2,
		VotesRequired: 2,
		Deadline:      time.Now().Add(10 * time.Second),
	})
	if err != nil {
		t.Fatalf("failed to create approval: %s", err)
	}

	appr, err := provider.approvalManager.Get("deployment/xxxx/dep-1:1.1.2")
	if err != nil {
		t.Fatalf("failed to get approval: %s", err)
	}
	if appr.Status() != types.ApprovalStatusApproved {
		t.Fatalf("approval not approved")
	}

	// creating "new version" event
	repo := types.Repository{
		Name: "gcr.io/v2-namespace/hello-world",
		Tag:  "1.1.2",
	}

	deps, err := provider.processEvent(&types.Event{Repository: repo})
	if err != nil {
		t.Errorf("failed to get deployments: %s", err)
	}

	if len(deps) != 1 {
		t.Errorf("expected to find 1 updated deployment but found %d", len(deps))
	}
}

func TestApprovalsCleanup(t *testing.T) {
	fp := &fakeImplementer{}
	fp.namespaces = &v1.NamespaceList{
		Items: []v1.Namespace{
			v1.Namespace{
				meta_v1.TypeMeta{},
				meta_v1.ObjectMeta{Name: "xxxx"},
				v1.NamespaceSpec{},
				v1.NamespaceStatus{},
			},
		},
	}
	deployments := []*apps_v1.Deployment{
		{
			meta_v1.TypeMeta{},
			meta_v1.ObjectMeta{
				Name:        "dep-1",
				Namespace:   "xxxx",
				Labels:      map[string]string{types.KeelPolicyLabel: "all", types.KeelMinimumApprovalsLabel: "1"},
				Annotations: map[string]string{},
			},
			apps_v1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							v1.Container{
								Image: "gcr.io/v2-namespace/hello-world:1.1.1",
							},
						},
					},
				},
			},
			apps_v1.DeploymentStatus{},
		},
	}
	grs := MustParseGRS(deployments)
	grc := &k8s.GenericResourceCache{}
	grc.Add(grs...)

	approver := approver()
	provider, err := NewProvider(fp, &fakeSender{}, approver, grc)
	if err != nil {
		t.Fatalf("failed to get provider: %s", err)
	}

	// approving event
	err = provider.approvalManager.Create(&types.Approval{
		Identifier:    "deployment/xxxx/dep-1:1.1.2",
		VotesReceived: 2,
		VotesRequired: 2,
		Deadline:      time.Now().Add(10 * time.Second),
	})
	if err != nil {
		t.Fatalf("failed to create approval: %s", err)
	}

	appr, err := provider.approvalManager.Get("deployment/xxxx/dep-1:1.1.2")
	if err != nil {
		t.Fatalf("failed to get approval: %s", err)
	}
	if appr.Status() != types.ApprovalStatusApproved {
		t.Fatalf("approval not approved")
	}

	// creating "new version" event
	repo := types.Repository{
		Name: "gcr.io/v2-namespace/hello-world",
		Tag:  "1.1.2",
	}

	deps, err := provider.processEvent(&types.Event{Repository: repo})
	if err != nil {
		t.Errorf("failed to get deployments: %s", err)
	}

	if len(deps) != 1 {
		t.Errorf("expected to find 1 updated deployment but found %d", len(deps))
	}

	// no approvals expected

	approvals, err := provider.approvalManager.List()
	if err != nil {
		t.Fatalf("failed to get a list of approvals: %s", err)
	}

	if len(approvals) != 0 {
		t.Errorf("expected to find 0 but found %d", len(approvals))
	}
}
