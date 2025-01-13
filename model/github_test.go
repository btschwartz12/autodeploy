package model

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/stretchr/testify/assert"
)

func TestParsePayload(t *testing.T) {
	req := httptest.NewRequest("POST", "/postreceive", bytes.NewBufferString(examplePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "push")

	hook, err := github.New()
	assert.NoError(t, err)

	payload, err := hook.Parse(req, github.PushEvent)
	assert.NoError(t, err)

	parsed, ok := payload.(github.PushPayload)
	assert.True(t, ok)

	pushEvent := PushEvent{}
	pushEvent.FromPayload(parsed)

	assert.Equal(t, "refs/heads/main", pushEvent.Ref)
	assert.Equal(t, "torvalds", pushEvent.Owner)
	assert.Equal(t, "linux", pushEvent.Repo)
	assert.Equal(t, "8e9703b922474b3d78aba29f388ea038396aab8d", pushEvent.AfterSha)
	assert.Equal(t, "53272ee3b33edfe8ba8db18881f25fb9a5234288", pushEvent.BeforeSha)
	assert.False(t, pushEvent.Forced)
}

const examplePayload = `
{
  "ref": "refs/heads/main",
  "before": "53272ee3b33edfe8ba8db18881f25fb9a5234288",
  "after": "8e9703b922474b3d78aba29f388ea038396aab8d",
  "repository": {
    "id": 914576987,
    "node_id": "R_kgDONoNWWw",
    "name": "linux",
    "full_name": "torvalds/linux",
    "private": true,
    "owner": {
      "name": "torvalds",
      "email": "105884019+torvalds@users.noreply.github.com",
      "login": "torvalds",
      "id": 105884019,
      "node_id": "U_kgDOBk-pcw",
      "avatar_url": "https://avatars.githubusercontent.com/u/105884019?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/torvalds",
      "html_url": "https://github.com/torvalds",
      "followers_url": "https://api.github.com/users/torvalds/followers",
      "following_url": "https://api.github.com/users/torvalds/following{/other_user}",
      "gists_url": "https://api.github.com/users/torvalds/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/torvalds/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/torvalds/subscriptions",
      "organizations_url": "https://api.github.com/users/torvalds/orgs",
      "repos_url": "https://api.github.com/users/torvalds/repos",
      "events_url": "https://api.github.com/users/torvalds/events{/privacy}",
      "received_events_url": "https://api.github.com/users/torvalds/received_events",
      "type": "User",
      "user_view_type": "public",
      "site_admin": false
    },
    "html_url": "https://github.com/torvalds/linux",
    "description": null,
    "fork": false,
    "url": "https://github.com/torvalds/linux",
    "forks_url": "https://api.github.com/repos/torvalds/linux/forks",
    "keys_url": "https://api.github.com/repos/torvalds/linux/keys{/key_id}",
    "collaborators_url": "https://api.github.com/repos/torvalds/linux/collaborators{/collaborator}",
    "teams_url": "https://api.github.com/repos/torvalds/linux/teams",
    "hooks_url": "https://api.github.com/repos/torvalds/linux/hooks",
    "issue_events_url": "https://api.github.com/repos/torvalds/linux/issues/events{/number}",
    "events_url": "https://api.github.com/repos/torvalds/linux/events",
    "assignees_url": "https://api.github.com/repos/torvalds/linux/assignees{/user}",
    "branches_url": "https://api.github.com/repos/torvalds/linux/branches{/branch}",
    "tags_url": "https://api.github.com/repos/torvalds/linux/tags",
    "blobs_url": "https://api.github.com/repos/torvalds/linux/git/blobs{/sha}",
    "git_tags_url": "https://api.github.com/repos/torvalds/linux/git/tags{/sha}",
    "git_refs_url": "https://api.github.com/repos/torvalds/linux/git/refs{/sha}",
    "trees_url": "https://api.github.com/repos/torvalds/linux/git/trees{/sha}",
    "statuses_url": "https://api.github.com/repos/torvalds/linux/statuses/{sha}",
    "languages_url": "https://api.github.com/repos/torvalds/linux/languages",
    "stargazers_url": "https://api.github.com/repos/torvalds/linux/stargazers",
    "contributors_url": "https://api.github.com/repos/torvalds/linux/contributors",
    "subscribers_url": "https://api.github.com/repos/torvalds/linux/subscribers",
    "subscription_url": "https://api.github.com/repos/torvalds/linux/subscription",
    "commits_url": "https://api.github.com/repos/torvalds/linux/commits{/sha}",
    "git_commits_url": "https://api.github.com/repos/torvalds/linux/git/commits{/sha}",
    "comments_url": "https://api.github.com/repos/torvalds/linux/comments{/number}",
    "issue_comment_url": "https://api.github.com/repos/torvalds/linux/issues/comments{/number}",
    "contents_url": "https://api.github.com/repos/torvalds/linux/contents/{+path}",
    "compare_url": "https://api.github.com/repos/torvalds/linux/compare/{base}...{head}",
    "merges_url": "https://api.github.com/repos/torvalds/linux/merges",
    "archive_url": "https://api.github.com/repos/torvalds/linux/{archive_format}{/ref}",
    "downloads_url": "https://api.github.com/repos/torvalds/linux/downloads",
    "issues_url": "https://api.github.com/repos/torvalds/linux/issues{/number}",
    "pulls_url": "https://api.github.com/repos/torvalds/linux/pulls{/number}",
    "milestones_url": "https://api.github.com/repos/torvalds/linux/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/torvalds/linux/notifications{?since,all,participating}",
    "labels_url": "https://api.github.com/repos/torvalds/linux/labels{/name}",
    "releases_url": "https://api.github.com/repos/torvalds/linux/releases{/id}",
    "deployments_url": "https://api.github.com/repos/torvalds/linux/deployments",
    "created_at": 1736459083,
    "updated_at": "2025-01-13T11:10:09Z",
    "pushed_at": 1736766631,
    "git_url": "git://github.com/torvalds/linux.git",
    "ssh_url": "git@github.com:torvalds/linux.git",
    "clone_url": "https://github.com/torvalds/linux.git",
    "svn_url": "https://github.com/torvalds/linux",
    "homepage": null,
    "size": 2,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": null,
    "has_issues": true,
    "has_projects": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": false,
    "has_discussions": false,
    "forks_count": 0,
    "mirror_url": null,
    "archived": false,
    "disabled": false,
    "open_issues_count": 0,
    "license": null,
    "allow_forking": true,
    "is_template": false,
    "web_commit_signoff_required": false,
    "topics": [],
    "visibility": "private",
    "forks": 0,
    "open_issues": 0,
    "watchers": 0,
    "default_branch": "main",
    "stargazers": 0,
    "master_branch": "main"
  },
  "pusher": {
    "name": "torvalds",
    "email": "105884019+torvalds@users.noreply.github.com"
  },
  "sender": {
    "login": "torvalds",
    "id": 105884019,
    "node_id": "U_kgDOBk-pcw",
    "avatar_url": "https://avatars.githubusercontent.com/u/105884019?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/torvalds",
    "html_url": "https://github.com/torvalds",
    "followers_url": "https://api.github.com/users/torvalds/followers",
    "following_url": "https://api.github.com/users/torvalds/following{/other_user}",
    "gists_url": "https://api.github.com/users/torvalds/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/torvalds/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/torvalds/subscriptions",
    "organizations_url": "https://api.github.com/users/torvalds/orgs",
    "repos_url": "https://api.github.com/users/torvalds/repos",
    "events_url": "https://api.github.com/users/torvalds/events{/privacy}",
    "received_events_url": "https://api.github.com/users/torvalds/received_events",
    "type": "User",
    "user_view_type": "public",
    "site_admin": false
  },
  "created": false,
  "deleted": false,
  "forced": false,
  "base_ref": null,
  "compare": "https://github.com/torvalds/linux/compare/53272ee3b33e...8e9703b92247",
  "commits": [
    {
      "id": "8e9703b922474b3d78aba29f388ea038396aab8d",
      "tree_id": "3ce6ba58c9196bfa5e88a954414f5d00f88c9b7a",
      "distinct": true,
      "message": "y",
      "timestamp": "2025-01-13T06:10:30-05:00",
      "url": "https://github.com/torvalds/linux/commit/8e9703b922474b3d78aba29f388ea038396aab8d",
      "author": {
        "name": "Ben Schwartz",
        "email": "scben@umich.edu",
        "username": "torvalds"
      },
      "committer": {
        "name": "Ben Schwartz",
        "email": "scben@umich.edu",
        "username": "torvalds"
      },
      "added": [],
      "removed": [],
      "modified": [
        "bruh"
      ]
    }
  ],
  "head_commit": {
    "id": "8e9703b922474b3d78aba29f388ea038396aab8d",
    "tree_id": "3ce6ba58c9196bfa5e88a954414f5d00f88c9b7a",
    "distinct": true,
    "message": "y",
    "timestamp": "2025-01-13T06:10:30-05:00",
    "url": "https://github.com/torvalds/linux/commit/8e9703b922474b3d78aba29f388ea038396aab8d",
    "author": {
      "name": "Ben Schwartz",
      "email": "scben@umich.edu",
      "username": "torvalds"
    },
    "committer": {
      "name": "Ben Schwartz",
      "email": "scben@umich.edu",
      "username": "torvalds"
    },
    "added": [],
    "removed": [],
    "modified": [
      "bruh"
    ]
  }
}
`
