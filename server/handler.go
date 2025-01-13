package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/btschwartz12/autodeploy/model"
	"github.com/go-playground/webhooks/v6/github"
)

const (
	timeout = 3 * time.Minute
)

var supportedEvents = []github.Event{
	github.PushEvent,
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := s.webhook.Parse(r, supportedEvents...)
	if err != nil {
		if err == github.ErrEventNotFound {
			s.logger.Infow("event not found")
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error parsing webhook", "error", err)
		http.Error(w, "error parsing webhook", http.StatusInternalServerError)
		return
	}

	err = s.handleEvent(payload)
	if err != nil {
		s.logger.Errorw("error handling event", "error", err)
		http.Error(w, "error handling event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleEvent(payload interface{}) error {
	switch event := payload.(type) {
	case github.PushPayload:
		pushEvent := model.PushEvent{}
		pushEvent.FromPayload(event)
		return s.handlePushEvent(&pushEvent)
	default:
		return fmt.Errorf("unsupported event type: %T", event)
	}
}

func (s *Server) handlePushEvent(event *model.PushEvent) error {
	s.logger.Infow("handling push event", "event", event)
	service := s.config.GetServiceByRepo(event.FullRepo())
	if service == nil {
		return fmt.Errorf("service not found for repo: %s", event.Repo)
	}
	go s.deployAsync(service, event)
	return nil
}

func (s *Server) deployAsync(service *model.Service, event *model.PushEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	go func() {
		defer cancel()
		err := s.deployer.Deploy(ctx, service, event)
		if ctx.Err() == context.DeadlineExceeded {
			s.slackClient.SendToSlack(getTimeoutMessage(service, event))
			s.logger.Errorw("deployment timeout", "service", service.Name)
			return
		}
		if err != nil {
			s.slackClient.SendToSlack(getFailureMessage(service, event, err))
			s.logger.Errorw("failed to deploy", "service", service.Name, "error", err)
		} else {
			s.slackClient.SendToSlack(getSuccessMessage(service, event))
			s.logger.Infow("deployed successfully", "service", service.Name)
		}
	}()
}

func getSuccessMessage(service *model.Service, event *model.PushEvent) (string, []string) {
	title := fmt.Sprintf("✅ successfully deployed `%s` ✅", service.Name)
	followUps := make([]string, 0)
	followUps = append(followUps, fmt.Sprintf("repo: `%s`", service.Repo))
	followUps = append(followUps, fmt.Sprintf("url: `%s`", service.HealthcheckURL))
	followUps = append(followUps, fmt.Sprintf("commit: `%s`", event.AfterSha))
	return title, followUps
}

func getFailureMessage(service *model.Service, event *model.PushEvent, err error) (string, []string) {
	title := fmt.Sprintf("❌ failed to deploy `%s` ❌", service.Name)
	followUps := make([]string, 0)
	followUps = append(followUps, fmt.Sprintf("repo: `%s`", service.Repo))
	followUps = append(followUps, fmt.Sprintf("commit: `%s`", event.AfterSha))
	followUps = append(followUps, fmt.Sprintf("error: \n```%s```", err.Error()))
	return title, followUps
}

func getTimeoutMessage(service *model.Service, event *model.PushEvent) (string, []string) {
	title := fmt.Sprintf("❌ deployment timeout for `%s` ❌", service.Name)
	followUps := make([]string, 0)
	followUps = append(followUps, fmt.Sprintf("repo: `%s`", service.Repo))
	followUps = append(followUps, fmt.Sprintf("commit: `%s`", event.AfterSha))
	return title, followUps
}
