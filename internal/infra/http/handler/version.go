package handler

import (
	"net/http"
	"strings"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/infra/http/middleware"
	"github.com/1995parham/koochooloo/internal/infra/http/response"
	"github.com/carlmjohnson/versioninfo"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// devel is reported when the binary carries no released version tag.
const devel = "devel"

// pseudoRevisionLen is the length of the commit-hash prefix Go embeds in a
// pseudo-version (e.g. v1.0.1-0.20260806171207-4b75e1e61a28).
const pseudoRevisionLen = 12

// release returns the released version tag the binary was built from, or
// "devel" when there is none. A pseudo-version is not a release: it only
// encodes the commit, which is reported separately (and to admins only).
func release(version, revision string) string {
	if version == "" || version == "unknown" || version == "(devel)" {
		return devel
	}

	// A pseudo-version ends in "-<12-char commit prefix>". Compare that exact
	// suffix against the embedded revision instead of scanning the whole
	// string, so a release tag that merely contains the prefix is not
	// mistaken for a pseudo-version.
	if _, commit, ok := strings.CutLast(version, "-"); ok &&
		len(revision) >= pseudoRevisionLen && commit == revision[:pseudoRevisionLen] {
		return devel
	}

	return version
}

// Version reports what the running binary was built from.
type Version struct {
	Logger *zap.Logger
	Tracer trace.Tracer
}

// Handle returns the version of the running binary. Admins additionally see
// the commit it was built from, when that commit was made and whether the
// working tree was dirty at build time.
func (h Version) Handle(c *echo.Context) error {
	_, span := h.Tracer.Start(c.Request().Context(), "handler.version")
	defer span.End()

	//nolint: exhaustruct_v5
	rs := response.Version{Version: release(versioninfo.Version, versioninfo.Revision)}

	claims, ok := middleware.ClaimsFrom(c)
	if ok && claims.Role.AtLeast(model.RoleAdmin) && versioninfo.Revision != "unknown" {
		rs.Revision = versioninfo.Revision
		rs.Dirty = versioninfo.DirtyBuild

		if !versioninfo.LastCommit.IsZero() {
			commit := versioninfo.LastCommit
			rs.LastCommit = &commit
		}
	}

	return c.JSON(http.StatusOK, rs)
}

// Register registers the routes of version handler on given echo group.
func (h Version) Register(g *echo.Group) {
	g.GET("/version", h.Handle)
}
