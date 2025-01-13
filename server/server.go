package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/webhooks/v6/github"
	"go.uber.org/zap"

	"github.com/btschwartz12/autodeploy/config"
	"github.com/btschwartz12/autodeploy/deploy"
	"github.com/btschwartz12/autodeploy/model"
	"github.com/btschwartz12/autodeploy/slack"
)

type Server struct {
	router      *chi.Mux
	logger      *zap.SugaredLogger
	slackClient *slack.SlackClient
	webhook     *github.Webhook
	deployer    *deploy.Deployer
	config      *model.Config
}

func NewServer(
	logger *zap.SugaredLogger,
	configPath string,
) (*Server, error) {

	c, err := config.New(configPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	h, err := github.New(github.Options.Secret(c.WebhookSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub webhook: %w", err)
	}

	s := &Server{
		logger:      logger,
		slackClient: slack.New(),
		webhook:     h,
		deployer:    deploy.New(logger, c.GithubToken),
		config:      c,
	}

	s.router = chi.NewRouter()
	s.router.Post(c.WebhookURLSuffix, s.handleWebhook)
	s.router.Get("/health", s.health)

	return s, nil
}

func (s *Server) GetRouter() *chi.Mux {
	return s.router
}
