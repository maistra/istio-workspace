package model_test

import (
	"fmt"
	"testing"

	"github.com/maistra/istio-workspace/pkg/model"
)

func TestLocatorStoreSort(t *testing.T) {
	store := model.LocatorStore{}
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "X", Kind: "X"}, Action: model.ActionCreate})
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "X", Kind: "X"}, Action: model.ActionDelete})
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "Y", Kind: "X"}, Action: model.ActionDelete})
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "Y", Kind: "X"}, Action: model.ActionCreate})

	ls := store.Store("X")
	for i, l := range ls {
		if (i == 0 || i == 1) && l.Action != model.ActionDelete {
			t.Error("should sort Delete first")
		}
		if (i == 2 || i == 3) && l.Action != model.ActionCreate {
			t.Error("should sort Delete first")
		}
		fmt.Println(l)
	}
}

func TestRefHash(t *testing.T) {
	r1 := model.Ref{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"X": "Y", "Y": "X"}}
	r2 := model.Ref{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}}

	refs := []model.Ref{
		{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "X"}},
		{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: false, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "Y", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("x"), Namespace: "X", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Remove: true},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X"},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y"},
		{KindName: model.ParseRefKindName("Y")},
	}

	if r1.Hash() != r2.Hash() {
		t.Errorf("Should match %v %v %v", r1.Hash() == r2.Hash(), r1.Hash(), r2.Hash())
	}

	for _, r := range refs {
		if r1.Hash() == r.Hash() {
			t.Errorf("Should match %v %v %v", r1.Hash() == r.Hash(), r1.Hash(), r.Hash())
		}
	}
}
