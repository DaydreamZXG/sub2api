package service

import "testing"

func TestBrandSiteName(t *testing.T) {
	cases := []struct {
		stored string
		want   string
	}{
		{stored: "", want: "YesChoy"},
		{stored: "   ", want: "YesChoy"},
		{stored: "Sub2API", want: "YesChoy"},
		{stored: " sub2api ", want: "YesChoy"},
		{stored: "My Station", want: "My Station"},
		{stored: " 野菜 AI ", want: "野菜 AI"},
	}
	for _, tc := range cases {
		if got := BrandSiteName(tc.stored); got != tc.want {
			t.Errorf("BrandSiteName(%q) = %q, want %q", tc.stored, got, tc.want)
		}
	}
}
