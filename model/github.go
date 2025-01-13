package model

import (
	"fmt"
	"strings"

	"github.com/go-playground/webhooks/v6/github"
)

type Commit struct {
	Sha       string `json:"sha"`
	Author    string `json:"author"`
	Committer string `json:"committer"`
	Message   string `json:"message"`
}

type PushEvent struct {
	Ref       string   `json:"ref"`
	BeforeSha string   `json:"before"`
	AfterSha  string   `json:"after"`
	Forced    bool     `json:"forced"`
	Pusher    string   `json:"pusher"`
	Owner     string   `json:"owner"`
	Repo      string   `json:"repo"`
	Commits   []Commit `json:"commits"`
}

func (p *PushEvent) FullRepo() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Repo)
}

func (p *PushEvent) FromPayload(payload github.PushPayload) {
	p.Ref = payload.Ref
	p.BeforeSha = payload.Before
	p.AfterSha = payload.After
	p.Forced = payload.Forced
	p.Pusher = payload.Pusher.Name
	parts := strings.Split(payload.Repository.FullName, "/")
	p.Owner = parts[0]
	p.Repo = parts[1]
	p.Commits = make([]Commit, len(payload.Commits))
	for i, commit := range payload.Commits {
		p.Commits[i] = Commit{
			Sha:       commit.ID,
			Author:    commit.Author.Username,
			Committer: commit.Committer.Username,
			Message:   commit.Message,
		}
	}
}
