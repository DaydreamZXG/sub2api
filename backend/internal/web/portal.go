package web

import "github.com/gin-gonic/gin"

// PortalPageFunc tells the frontend server whether a request arrived on the resellers' client domain
// and, if so, which config to put into the page instead of the site's public settings. The config is
// JSON with at least "portal": true; "site_name" and "site_logo" also set the tab's title and icon.
type PortalPageFunc func(c *gin.Context) (configJSON []byte, ok bool)
