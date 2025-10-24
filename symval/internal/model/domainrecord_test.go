package model

import "testing"

func TestGroupByGroupID(t *testing.T) {
	records := []*DomainRecord{
		{GroupID: "group1", Hostname: "a.com"},
		{GroupID: "group1", Hostname: "b.com"},
		{GroupID: "group2", Hostname: "c.com"},
	}

	grouped := GroupByGroupID(records)
	if len(grouped) != 2 {
		t.Errorf("expected 2 groups, got %d", len(grouped))
	}
	if len(grouped["group1"]) != 2 {
		t.Errorf("expected 2 records in group1, got %d", len(grouped["group1"]))
	}
	if len(grouped["group2"]) != 1 {
		t.Errorf("expected 1 record in group2, got %d", len(grouped["group2"]))
	}
}
