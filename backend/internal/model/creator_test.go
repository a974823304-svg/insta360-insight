package model

import (
	"encoding/json"
	"testing"
)

func TestAudienceJSONShape(t *testing.T) {
	a := Audience{
		Age:    []AgeShare{{Bucket: "25-34 岁", Share: 42.7, Color: "#FFB547"}},
		Gender: []GenderShare{{Gender: "男", Share: 58.4, Color: "#5EA1FF"}},
	}
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Audience
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Age) != 1 || len(out.Gender) != 1 {
		t.Fatalf("期望 age/gender 各 1 条, 实际 %d/%d", len(out.Age), len(out.Gender))
	}
	if out.Gender[0].Gender != "男" {
		t.Fatalf("Gender 字段丢失: %+v", out.Gender[0])
	}
}
