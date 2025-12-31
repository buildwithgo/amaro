package routers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/buildwithgo/amaro"
)

// Benchmark routing performance
func BenchmarkTrieRouter_Static(b *testing.B) {
	r := NewTrieRouter()
	handler := func(c *amaro.Context) error { return nil }
	r.GET("/hello", handler)
	r.GET("/users/list", handler)
	r.GET("/api/v1/status", handler)

	ctx := amaro.NewContext(nil, nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Find(http.MethodGet, "/users/list", ctx)
	}
}

func BenchmarkTrieRouter_Param(b *testing.B) {
	r := NewTrieRouter()
	handler := func(c *amaro.Context) error { return nil }
	r.GET("/users/:id", handler)
	r.GET("/users/:id/posts/:post_id", handler)

	ctx := amaro.NewContext(nil, nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx.Reset(nil, nil) // Reset params
		_, _ = r.Find(http.MethodGet, "/users/123/posts/456", ctx)
	}
}

func BenchmarkTrieRouter_Wildcard(b *testing.B) {
	r := NewTrieRouter()
	handler := func(c *amaro.Context) error { return nil }
	r.GET("/static/*filepath", handler)

	ctx := amaro.NewContext(nil, nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx.Reset(nil, nil)
		_, _ = r.Find(http.MethodGet, "/static/css/main.css", ctx)
	}
}

func BenchmarkTrieRouter_GithubAPI(b *testing.B) {
	// Simulate GitHub API structure
	r := NewTrieRouter()
	handler := func(c *amaro.Context) error { return nil }

	routes := []string{
		"/authorizations",
		"/authorizations/:id",
		"/applications/:client_id/tokens/:access_token",
		"/events",
		"/repos/:owner/:repo/events",
		"/networks/:owner/:repo/events",
		"/orgs/:org/events",
		"/users/:user/received_events",
		"/users/:user/received_events/public",
		"/users/:user/events",
		"/users/:user/events/public",
		"/users/:user/events/orgs/:org",
		"/feeds",
		"/notifications",
		"/notifications/threads/:id/subscription",
		"/repos/:owner/:repo/notifications",
		"/repos/:owner/:repo/stargazers",
		"/users/:user/starred",
		"/users/:user/starred/:owner/:repo",
		"/repos/:owner/:repo/subscribers",
		"/users/:user/subscriptions",
		"/users/:user/subscriptions/:owner/:repo",
		"/user/subscriptions",
		"/user/subscriptions/:owner/:repo",
		"/users/:user/gists",
		"/gists",
		"/gists/:id",
		"/gists/:id/star",
		"/repos/:owner/:repo/git/blobs/:sha",
		"/repos/:owner/:repo/git/commits/:sha",
		"/repos/:owner/:repo/git/refs",
		"/repos/:owner/:repo/git/tags/:sha",
		"/repos/:owner/:repo/git/trees/:sha",
		"/issues",
		"/user/issues",
		"/orgs/:org/issues",
		"/repos/:owner/:repo/issues",
		"/repos/:owner/:repo/issues/:number",
		"/repos/:owner/:repo/issues/:number/lock",
		"/repos/:owner/:repo/assignees",
		"/repos/:owner/:repo/assignees/:assignee",
		"/repos/:owner/:repo/issues/:number/comments",
		"/repos/:owner/:repo/issues/comments",
		"/repos/:owner/:repo/issues/comments/:id",
		"/repos/:owner/:repo/labels",
		"/repos/:owner/:repo/labels/:name",
		"/repos/:owner/:repo/issues/:number/labels",
		"/repos/:owner/:repo/milestones/:number/labels",
		"/repos/:owner/:repo/milestones",
		"/repos/:owner/:repo/milestones/:number",
		"/emojis",
		"/gitignore/templates",
		"/gitignore/templates/:name",
		"/meta",
		"/rate_limit",
		"/users/:user/orgs",
		"/user/orgs",
		"/orgs/:org",
		"/orgs/:org/members",
		"/orgs/:org/members/:user",
		"/orgs/:org/public_members",
		"/orgs/:org/public_members/:user",
		"/orgs/:org/teams",
		"/teams/:id",
		"/teams/:id/members",
		"/teams/:id/members/:user",
		"/teams/:id/repos",
		"/teams/:id/repos/:owner/:repo",
		"/user/teams",
		"/repos/:owner/:repo/pulls",
		"/repos/:owner/:repo/pulls/:number",
		"/repos/:owner/:repo/pulls/:number/commits",
		"/repos/:owner/:repo/pulls/:number/files",
		"/repos/:owner/:repo/pulls/:number/merge",
		"/repos/:owner/:repo/pulls/:number/comments",
		"/repos/:owner/:repo/pulls/comments",
		"/repos/:owner/:repo/pulls/comments/:number",
		"/repos/:owner/:repo",
		"/repos/:owner/:repo/contributors",
		"/repos/:owner/:repo/languages",
		"/repos/:owner/:repo/teams",
		"/repos/:owner/:repo/tags",
		"/repos/:owner/:repo/branches",
		"/repos/:owner/:repo/branches/:branch",
		"/repos/:owner/:repo/collaborators",
		"/repos/:owner/:repo/collaborators/:user",
		"/repos/:owner/:repo/comments",
		"/repos/:owner/:repo/comments/:id",
		"/repos/:owner/:repo/commits",
		"/repos/:owner/:repo/commits/:sha",
		"/repos/:owner/:repo/commits/:sha/comments",
		"/repos/:owner/:repo/keys",
		"/repos/:owner/:repo/keys/:id",
		"/repos/:owner/:repo/contents/*path",
		"/repos/:owner/:repo/downloads",
		"/repos/:owner/:repo/downloads/:id",
		"/repos/:owner/:repo/forks",
		"/repos/:owner/:repo/hooks",
		"/repos/:owner/:repo/hooks/:id",
		"/repos/:owner/:repo/releases",
		"/repos/:owner/:repo/releases/:id",
		"/repos/:owner/:repo/releases/:id/assets",
		"/repos/:owner/:repo/stats/contributors",
		"/repos/:owner/:repo/stats/commit_activity",
		"/repos/:owner/:repo/stats/code_frequency",
		"/repos/:owner/:repo/stats/participation",
		"/repos/:owner/:repo/stats/punch_card",
		"/repos/:owner/:repo/statuses/:ref",
		"/search/repositories",
		"/search/code",
		"/search/issues",
		"/search/users",
		"/users/:user",
		"/user",
		"/users",
		"/user/emails",
		"/user/followers",
		"/user/following",
		"/user/following/:user",
		"/users/:user/followers",
		"/users/:user/following",
		"/users/:user/following/:target_user",
		"/users/:user/keys",
		"/users/:user/keys/:id",
		"/user/keys",
		"/user/keys/:id",
	}

	for _, route := range routes {
		if err := r.GET(route, handler); err != nil {
			panic(fmt.Sprintf("failed to register route %s: %v", route, err))
		}
	}

	ctx := amaro.NewContext(nil, nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Test a deep route
		ctx.Reset(nil, nil)
		_, _ = r.Find(http.MethodGet, "/repos/octocat/hello-world/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e/comments", ctx)
	}
}
