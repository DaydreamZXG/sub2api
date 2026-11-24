package service

import "strings"

const (
	brandSiteName = "YesChoy"
	// The name a fresh install stores in the site_name setting.
	stockSiteName = "Sub2API"
)

// BrandSiteName returns the name to show for a stored site name: an empty value and the
// stock name of a fresh install both mean "not set", so the brand name is used instead.
func BrandSiteName(stored string) string {
	name := strings.TrimSpace(stored)
	if name == "" || strings.EqualFold(name, stockSiteName) {
		return brandSiteName
	}
	return name
}
