package model

import (
	"encoding/json"
	"testing"
)

func TestAnalysisModelJSONShape(t *testing.T) {
	in := struct {
		C ContentItem  `json:"c"`
		M Competitor   `json:"m"`
		P PartnerBrand `json:"p"`
		T TagItem      `json:"t"`
	}{
		C: ContentItem{ID: 1, Title: "滑雪教程", Form: "教程", Topic: "滑雪", Views: 120000, Engagement: 7.3, IsHit: true},
		M: Competitor{Name: "GoPro", Category: "运动相机", Buzz: 980000, Growth: 12.4, Sentiment: 78.0},
		P: PartnerBrand{Name: "红牛", Industry: "饮料", Contents: 8, Exposure: 5400000, Engagement: 320000, ROI: 4.2},
		T: TagItem{Word: "画质", Weight: 86.0},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out struct {
		C ContentItem  `json:"c"`
		M Competitor   `json:"m"`
		P PartnerBrand `json:"p"`
		T TagItem      `json:"t"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.C.Title != "滑雪教程" || !out.C.IsHit {
		t.Fatalf("ContentItem 字段丢失: %+v", out.C)
	}
	if out.M.Name != "GoPro" || out.M.Buzz != 980000 {
		t.Fatalf("Competitor 字段丢失: %+v", out.M)
	}
	if out.P.ROI != 4.2 {
		t.Fatalf("PartnerBrand.ROI 丢失: %+v", out.P)
	}
	if out.T.Word != "画质" {
		t.Fatalf("TagItem 字段丢失: %+v", out.T)
	}
}
