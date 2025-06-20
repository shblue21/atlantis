package webhooks_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultMSTeamsClient_PostMessage(t *testing.T) {
	t.Run("should send correct message format", func(t *testing.T) {
		var receivedMessage webhooks.TeamsMessage
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}
			
			err := json.NewDecoder(r.Body).Decode(&receivedMessage)
			if err != nil {
				t.Errorf("Failed to decode message: %v", err)
			}
			
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := webhooks.NewMSTeamsClient()
		
		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Repo: models.Repo{
				FullName: "owner/repo",
			},
			Pull: models.PullRequest{
				BaseBranch: "main",
				Num:        1,
				URL:        "https://github.com/owner/repo/pull/1",
			},
			User: models.User{
				Username: "atlantis",
			},
			Success:     true,
			Directory:   ".",
			ProjectName: "test-project",
		}

		err := client.PostMessage(server.URL, applyResult)
		Ok(t, err)

		// Verify message structure
		Equals(t, "MessageCard", receivedMessage.Type)
		Equals(t, "http://schema.org/extensions", receivedMessage.Context)
		Equals(t, "good", receivedMessage.ThemeColor)
		Equals(t, "Terraform apply succeeded for owner/repo", receivedMessage.Summary)
		
		// Verify sections
		Assert(t, len(receivedMessage.Sections) == 1, "Expected 1 section")
		section := receivedMessage.Sections[0]
		Equals(t, "Atlantis Apply succeeded", section.ActivityTitle)
		Equals(t, "Repository: owner/repo", section.ActivitySubtitle)
		Assert(t, section.Markdown, "Expected markdown to be true")
		
		// Verify facts
		expectedFacts := map[string]string{
			"Workspace":    "production",
			"Branch":       "main",
			"User":         "atlantis",
			"Directory":    "/",
			"Pull Request": "[#1](https://github.com/owner/repo/pull/1)",
			"Project":      "test-project",
		}
		
		Assert(t, len(section.Facts) == len(expectedFacts), "Expected %d facts, got %d", len(expectedFacts), len(section.Facts))
		
		for _, fact := range section.Facts {
			expectedValue, exists := expectedFacts[fact.Name]
			Assert(t, exists, "Unexpected fact: %s", fact.Name)
			Equals(t, expectedValue, fact.Value)
		}
	})

	t.Run("should handle failure case", func(t *testing.T) {
		var receivedMessage webhooks.TeamsMessage
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&receivedMessage)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := webhooks.NewMSTeamsClient()
		
		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Repo: models.Repo{
				FullName: "owner/repo",
			},
			Pull: models.PullRequest{
				BaseBranch: "main",
				Num:        1,
				URL:        "https://github.com/owner/repo/pull/1",
			},
			User: models.User{
				Username: "atlantis",
			},
			Success:   false,
			Directory: "terraform/",
		}

		err := client.PostMessage(server.URL, applyResult)
		Ok(t, err)

		// Verify failure message
		Equals(t, "attention", receivedMessage.ThemeColor)
		Equals(t, "Terraform apply failed for owner/repo", receivedMessage.Summary)
		Equals(t, "Atlantis Apply failed", receivedMessage.Sections[0].ActivityTitle)
	})

	t.Run("should handle server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := webhooks.NewMSTeamsClient()
		
		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Success:   true,
		}

		err := client.PostMessage(server.URL, applyResult)
		Assert(t, err != nil, "Expected error for server error response")
	})
}
