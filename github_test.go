package mux

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/nikandfor/assert"
)

// Routers are taken from https://github.com/gin-gonic/gin

type route struct {
	method string
	path   string
}

// http://developer.github.com/v3/
var githubAPI = []route{
	// OAuth Authorizations
	{http.MethodGet, "/authorizations"},
	{http.MethodGet, "/authorizations/:id"},
	{http.MethodPost, "/authorizations"},
	//{http.MethodPut, "/authorizations/clients/:client_id"},
	//{http.MethodPatch, "/authorizations/:id"},
	{http.MethodDelete, "/authorizations/:id"},
	{http.MethodGet, "/applications/:client_id/tokens/:access_token"},
	{http.MethodDelete, "/applications/:client_id/tokens"},
	{http.MethodDelete, "/applications/:client_id/tokens/:access_token"},

	// Activity
	{http.MethodGet, "/events"},
	{http.MethodGet, "/repos/:owner/:repo/events"},
	{http.MethodGet, "/networks/:owner/:repo/events"},
	{http.MethodGet, "/orgs/:org/events"},
	{http.MethodGet, "/users/:user/received_events"},
	{http.MethodGet, "/users/:user/received_events/public"},
	{http.MethodGet, "/users/:user/events"},
	{http.MethodGet, "/users/:user/events/public"},
	{http.MethodGet, "/users/:user/events/orgs/:org"},
	{http.MethodGet, "/feeds"},
	{http.MethodGet, "/notifications"},
	{http.MethodGet, "/repos/:owner/:repo/notifications"},
	{http.MethodPut, "/notifications"},
	{http.MethodPut, "/repos/:owner/:repo/notifications"},
	{http.MethodGet, "/notifications/threads/:id"},
	//{http.MethodPatch, "/notifications/threads/:id"},
	{http.MethodGet, "/notifications/threads/:id/subscription"},
	{http.MethodPut, "/notifications/threads/:id/subscription"},
	{http.MethodDelete, "/notifications/threads/:id/subscription"},
	{http.MethodGet, "/repos/:owner/:repo/stargazers"},
	{http.MethodGet, "/users/:user/starred"},
	{http.MethodGet, "/user/starred"},
	{http.MethodGet, "/user/starred/:owner/:repo"},
	{http.MethodPut, "/user/starred/:owner/:repo"},
	{http.MethodDelete, "/user/starred/:owner/:repo"},
	{http.MethodGet, "/repos/:owner/:repo/subscribers"},
	{http.MethodGet, "/users/:user/subscriptions"},
	{http.MethodGet, "/user/subscriptions"},
	{http.MethodGet, "/repos/:owner/:repo/subscription"},
	{http.MethodPut, "/repos/:owner/:repo/subscription"},
	{http.MethodDelete, "/repos/:owner/:repo/subscription"},
	{http.MethodGet, "/user/subscriptions/:owner/:repo"},
	{http.MethodPut, "/user/subscriptions/:owner/:repo"},
	{http.MethodDelete, "/user/subscriptions/:owner/:repo"},

	// Gists
	{http.MethodGet, "/users/:user/gists"},
	{http.MethodGet, "/gists"},
	//{http.MethodGet, "/gists/public"},
	//{http.MethodGet, "/gists/starred"},
	{http.MethodGet, "/gists/:id"},
	{http.MethodPost, "/gists"},
	//{http.MethodPatch, "/gists/:id"},
	{http.MethodPut, "/gists/:id/star"},
	{http.MethodDelete, "/gists/:id/star"},
	{http.MethodGet, "/gists/:id/star"},
	{http.MethodPost, "/gists/:id/forks"},
	{http.MethodDelete, "/gists/:id"},

	// Git Data
	{http.MethodGet, "/repos/:owner/:repo/git/blobs/:sha"},
	{http.MethodPost, "/repos/:owner/:repo/git/blobs"},
	{http.MethodGet, "/repos/:owner/:repo/git/commits/:sha"},
	{http.MethodPost, "/repos/:owner/:repo/git/commits"},
	//{http.MethodGet, "/repos/:owner/:repo/git/refs/*ref"},
	{http.MethodGet, "/repos/:owner/:repo/git/refs"},
	{http.MethodPost, "/repos/:owner/:repo/git/refs"},
	//{http.MethodPatch, "/repos/:owner/:repo/git/refs/*ref"},
	//{http.MethodDelete, "/repos/:owner/:repo/git/refs/*ref"},
	{http.MethodGet, "/repos/:owner/:repo/git/tags/:sha"},
	{http.MethodPost, "/repos/:owner/:repo/git/tags"},
	{http.MethodGet, "/repos/:owner/:repo/git/trees/:sha"},
	{http.MethodPost, "/repos/:owner/:repo/git/trees"},

	// Issues
	{http.MethodGet, "/issues"},
	{http.MethodGet, "/user/issues"},
	{http.MethodGet, "/orgs/:org/issues"},
	{http.MethodGet, "/repos/:owner/:repo/issues"},
	{http.MethodGet, "/repos/:owner/:repo/issues/:number"},
	{http.MethodPost, "/repos/:owner/:repo/issues"},
	//{http.MethodPatch, "/repos/:owner/:repo/issues/:number"},
	{http.MethodGet, "/repos/:owner/:repo/assignees"},
	{http.MethodGet, "/repos/:owner/:repo/assignees/:assignee"},
	{http.MethodGet, "/repos/:owner/:repo/issues/:number/comments"},
	//{http.MethodGet, "/repos/:owner/:repo/issues/comments"},
	//{http.MethodGet, "/repos/:owner/:repo/issues/comments/:id"},
	{http.MethodPost, "/repos/:owner/:repo/issues/:number/comments"},
	//{http.MethodPatch, "/repos/:owner/:repo/issues/comments/:id"},
	//{http.MethodDelete, "/repos/:owner/:repo/issues/comments/:id"},
	{http.MethodGet, "/repos/:owner/:repo/issues/:number/events"},
	//{http.MethodGet, "/repos/:owner/:repo/issues/events"},
	//{http.MethodGet, "/repos/:owner/:repo/issues/events/:id"},
	{http.MethodGet, "/repos/:owner/:repo/labels"},
	{http.MethodGet, "/repos/:owner/:repo/labels/:name"},
	{http.MethodPost, "/repos/:owner/:repo/labels"},
	//{http.MethodPatch, "/repos/:owner/:repo/labels/:name"},
	{http.MethodDelete, "/repos/:owner/:repo/labels/:name"},
	{http.MethodGet, "/repos/:owner/:repo/issues/:number/labels"},
	{http.MethodPost, "/repos/:owner/:repo/issues/:number/labels"},
	{http.MethodDelete, "/repos/:owner/:repo/issues/:number/labels/:name"},
	{http.MethodPut, "/repos/:owner/:repo/issues/:number/labels"},
	{http.MethodDelete, "/repos/:owner/:repo/issues/:number/labels"},
	{http.MethodGet, "/repos/:owner/:repo/milestones/:number/labels"},
	{http.MethodGet, "/repos/:owner/:repo/milestones"},
	{http.MethodGet, "/repos/:owner/:repo/milestones/:number"},
	{http.MethodPost, "/repos/:owner/:repo/milestones"},
	//{http.MethodPatch, "/repos/:owner/:repo/milestones/:number"},
	{http.MethodDelete, "/repos/:owner/:repo/milestones/:number"},

	// Miscellaneous
	{http.MethodGet, "/emojis"},
	{http.MethodGet, "/gitignore/templates"},
	{http.MethodGet, "/gitignore/templates/:name"},
	{http.MethodPost, "/markdown"},
	{http.MethodPost, "/markdown/raw"},
	{http.MethodGet, "/meta"},
	{http.MethodGet, "/rate_limit"},

	// Organizations
	{http.MethodGet, "/users/:user/orgs"},
	{http.MethodGet, "/user/orgs"},
	{http.MethodGet, "/orgs/:org"},
	//{http.MethodPatch, "/orgs/:org"},
	{http.MethodGet, "/orgs/:org/members"},
	{http.MethodGet, "/orgs/:org/members/:user"},
	{http.MethodDelete, "/orgs/:org/members/:user"},
	{http.MethodGet, "/orgs/:org/public_members"},
	{http.MethodGet, "/orgs/:org/public_members/:user"},
	{http.MethodPut, "/orgs/:org/public_members/:user"},
	{http.MethodDelete, "/orgs/:org/public_members/:user"},
	{http.MethodGet, "/orgs/:org/teams"},
	{http.MethodGet, "/teams/:id"},
	{http.MethodPost, "/orgs/:org/teams"},
	//{http.MethodPatch, "/teams/:id"},
	{http.MethodDelete, "/teams/:id"},
	{http.MethodGet, "/teams/:id/members"},
	{http.MethodGet, "/teams/:id/members/:user"},
	{http.MethodPut, "/teams/:id/members/:user"},
	{http.MethodDelete, "/teams/:id/members/:user"},
	{http.MethodGet, "/teams/:id/repos"},
	{http.MethodGet, "/teams/:id/repos/:owner/:repo"},
	{http.MethodPut, "/teams/:id/repos/:owner/:repo"},
	{http.MethodDelete, "/teams/:id/repos/:owner/:repo"},
	{http.MethodGet, "/user/teams"},

	// Pull Requests
	{http.MethodGet, "/repos/:owner/:repo/pulls"},
	{http.MethodGet, "/repos/:owner/:repo/pulls/:number"},
	{http.MethodPost, "/repos/:owner/:repo/pulls"},
	//{http.MethodPatch, "/repos/:owner/:repo/pulls/:number"},
	{http.MethodGet, "/repos/:owner/:repo/pulls/:number/commits"},
	{http.MethodGet, "/repos/:owner/:repo/pulls/:number/files"},
	{http.MethodGet, "/repos/:owner/:repo/pulls/:number/merge"},
	{http.MethodPut, "/repos/:owner/:repo/pulls/:number/merge"},
	{http.MethodGet, "/repos/:owner/:repo/pulls/:number/comments"},
	//{http.MethodGet, "/repos/:owner/:repo/pulls/comments"},
	//{http.MethodGet, "/repos/:owner/:repo/pulls/comments/:number"},
	{http.MethodPut, "/repos/:owner/:repo/pulls/:number/comments"},
	//{http.MethodPatch, "/repos/:owner/:repo/pulls/comments/:number"},
	//{http.MethodDelete, "/repos/:owner/:repo/pulls/comments/:number"},

	// Repositories
	{http.MethodGet, "/user/repos"},
	{http.MethodGet, "/users/:user/repos"},
	{http.MethodGet, "/orgs/:org/repos"},
	{http.MethodGet, "/repositories"},
	{http.MethodPost, "/user/repos"},
	{http.MethodPost, "/orgs/:org/repos"},
	{http.MethodGet, "/repos/:owner/:repo"},
	//{http.MethodPatch, "/repos/:owner/:repo"},
	{http.MethodGet, "/repos/:owner/:repo/contributors"},
	{http.MethodGet, "/repos/:owner/:repo/languages"},
	{http.MethodGet, "/repos/:owner/:repo/teams"},
	{http.MethodGet, "/repos/:owner/:repo/tags"},
	{http.MethodGet, "/repos/:owner/:repo/branches"},
	{http.MethodGet, "/repos/:owner/:repo/branches/:branch"},
	{http.MethodDelete, "/repos/:owner/:repo"},
	{http.MethodGet, "/repos/:owner/:repo/collaborators"},
	{http.MethodGet, "/repos/:owner/:repo/collaborators/:user"},
	{http.MethodPut, "/repos/:owner/:repo/collaborators/:user"},
	{http.MethodDelete, "/repos/:owner/:repo/collaborators/:user"},
	{http.MethodGet, "/repos/:owner/:repo/comments"},
	{http.MethodGet, "/repos/:owner/:repo/commits/:sha/comments"},
	{http.MethodPost, "/repos/:owner/:repo/commits/:sha/comments"},
	{http.MethodGet, "/repos/:owner/:repo/comments/:id"},
	//{http.MethodPatch, "/repos/:owner/:repo/comments/:id"},
	{http.MethodDelete, "/repos/:owner/:repo/comments/:id"},
	{http.MethodGet, "/repos/:owner/:repo/commits"},
	{http.MethodGet, "/repos/:owner/:repo/commits/:sha"},
	{http.MethodGet, "/repos/:owner/:repo/readme"},
	//{http.MethodGet, "/repos/:owner/:repo/contents/*path"},
	//{http.MethodPut, "/repos/:owner/:repo/contents/*path"},
	//{http.MethodDelete, "/repos/:owner/:repo/contents/*path"},
	//{http.MethodGet, "/repos/:owner/:repo/:archive_format/:ref"},
	{http.MethodGet, "/repos/:owner/:repo/keys"},
	{http.MethodGet, "/repos/:owner/:repo/keys/:id"},
	{http.MethodPost, "/repos/:owner/:repo/keys"},
	//{http.MethodPatch, "/repos/:owner/:repo/keys/:id"},
	{http.MethodDelete, "/repos/:owner/:repo/keys/:id"},
	{http.MethodGet, "/repos/:owner/:repo/downloads"},
	{http.MethodGet, "/repos/:owner/:repo/downloads/:id"},
	{http.MethodDelete, "/repos/:owner/:repo/downloads/:id"},
	{http.MethodGet, "/repos/:owner/:repo/forks"},
	{http.MethodPost, "/repos/:owner/:repo/forks"},
	{http.MethodGet, "/repos/:owner/:repo/hooks"},
	{http.MethodGet, "/repos/:owner/:repo/hooks/:id"},
	{http.MethodPost, "/repos/:owner/:repo/hooks"},
	//{http.MethodPatch, "/repos/:owner/:repo/hooks/:id"},
	{http.MethodPost, "/repos/:owner/:repo/hooks/:id/tests"},
	{http.MethodDelete, "/repos/:owner/:repo/hooks/:id"},
	{http.MethodPost, "/repos/:owner/:repo/merges"},
	{http.MethodGet, "/repos/:owner/:repo/releases"},
	{http.MethodGet, "/repos/:owner/:repo/releases/:id"},
	{http.MethodPost, "/repos/:owner/:repo/releases"},
	//{http.MethodPatch, "/repos/:owner/:repo/releases/:id"},
	{http.MethodDelete, "/repos/:owner/:repo/releases/:id"},
	{http.MethodGet, "/repos/:owner/:repo/releases/:id/assets"},
	{http.MethodGet, "/repos/:owner/:repo/stats/contributors"},
	{http.MethodGet, "/repos/:owner/:repo/stats/commit_activity"},
	{http.MethodGet, "/repos/:owner/:repo/stats/code_frequency"},
	{http.MethodGet, "/repos/:owner/:repo/stats/participation"},
	{http.MethodGet, "/repos/:owner/:repo/stats/punch_card"},
	{http.MethodGet, "/repos/:owner/:repo/statuses/:ref"},
	{http.MethodPost, "/repos/:owner/:repo/statuses/:ref"},

	// Search
	{http.MethodGet, "/search/repositories"},
	{http.MethodGet, "/search/code"},
	{http.MethodGet, "/search/issues"},
	{http.MethodGet, "/search/users"},
	{http.MethodGet, "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{http.MethodGet, "/legacy/repos/search/:keyword"},
	{http.MethodGet, "/legacy/user/search/:keyword"},
	{http.MethodGet, "/legacy/user/email/:email"},

	// Users
	{http.MethodGet, "/users/:user"},
	{http.MethodGet, "/user"},
	//{http.MethodPatch, "/user"},
	{http.MethodGet, "/users"},
	{http.MethodGet, "/user/emails"},
	{http.MethodPost, "/user/emails"},
	{http.MethodDelete, "/user/emails"},
	{http.MethodGet, "/users/:user/followers"},
	{http.MethodGet, "/user/followers"},
	{http.MethodGet, "/users/:user/following"},
	{http.MethodGet, "/user/following"},
	{http.MethodGet, "/user/following/:user"},
	{http.MethodGet, "/users/:user/following/:target_user"},
	{http.MethodPut, "/user/following/:user"},
	{http.MethodDelete, "/user/following/:user"},
	{http.MethodGet, "/users/:user/keys"},
	{http.MethodGet, "/user/keys"},
	{http.MethodGet, "/user/keys/:id"},
	{http.MethodPost, "/user/keys"},
	//{http.MethodPatch, "/user/keys/:id"},
	{http.MethodDelete, "/user/keys/:id"},
}

var dumpAll = flag.Bool("dump-all", false, "dump tree while tested")

func TestGithub(t *testing.T) {
	var hpath, hreq string
	h := func(p string) HandlerFunc {
		return HandlerFunc(func(c *Context) error {
			hpath = p
			hreq = c.Request.RequestURI
			return nil
		})
	}

	replace := map[string]string{}

	req := &http.Request{}

	m := New()

	for i, tc := range githubAPI {
		replace[tc.path] = strings.ReplaceAll(tc.path, ":", "")

		if *dumpAll {
			fmt.Printf("ADD HANDLER: %v %v\n", tc.method, tc.path)
		}

		m.Handle(tc.method, tc.path, h(tc.path))

		for _, tc := range githubAPI[:i+1] {
			func() {
				c := NewContext(nil, req)
				defer FreeContext(c)

				req.RequestURI = replace[tc.path]

				h := m.Lookup(tc.method, req.RequestURI, c)
				if assert.True(t, h != nil, "%v %v", tc.method, req.RequestURI) {
					err := h(c)
					assert.NoError(t, err)
					assert.Equal(t, tc.path, hpath)
					assert.Equal(t, req.RequestURI, hreq)
				}
			}()

			if t.Failed() {
				break
			}
		}

		if *dumpAll {
			for meth, root := range m.meth {
				fmt.Printf("routes dump %v\n%s\n", meth, m.dumpString(root))
			}
		}

		if t.Failed() {
			break
		}
	}
}
