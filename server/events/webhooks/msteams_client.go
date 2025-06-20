// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	teamsSuccessColor = "good"
	teamsFailureColor = "attention"
)

//go:generate pegomock generate --package mocks -o mocks/mock_msteams_client.go MSTeamsClient

// MSTeamsClient handles making API calls to Microsoft Teams.
type MSTeamsClient interface {
	PostMessage(webhookURL string, applyResult ApplyResult) error
}

// DefaultMSTeamsClient is the default implementation of MSTeamsClient.
type DefaultMSTeamsClient struct {
	Client *http.Client
}

// NewMSTeamsClient creates a new MS Teams client.
func NewMSTeamsClient() MSTeamsClient {
	return &DefaultMSTeamsClient{
		Client: http.DefaultClient,
	}
}

// PostMessage sends a message to MS Teams using the webhook URL.
func (d *DefaultMSTeamsClient) PostMessage(webhookURL string, applyResult ApplyResult) error {
	message := d.createMessage(applyResult)
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshaling teams message: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("teams webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// TeamsMessage represents the structure of a Microsoft Teams message card.
type TeamsMessage struct {
	Type       string                `json:"@type"`
	Context    string                `json:"@context"`
	ThemeColor string                `json:"themeColor"`
	Summary    string                `json:"summary"`
	Sections   []TeamsMessageSection `json:"sections"`
}

// TeamsMessageSection represents a section in a Teams message card.
type TeamsMessageSection struct {
	ActivityTitle    string              `json:"activityTitle"`
	ActivitySubtitle string              `json:"activitySubtitle,omitempty"`
	ActivityImage    string              `json:"activityImage,omitempty"`
	Facts            []TeamsMessageFact  `json:"facts,omitempty"`
	Markdown         bool                `json:"markdown"`
}

// TeamsMessageFact represents a fact in a Teams message section.
type TeamsMessageFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// createMessage creates a Teams message from the apply result.
func (d *DefaultMSTeamsClient) createMessage(applyResult ApplyResult) TeamsMessage {
	var color string
	var successWord string
	var activityImage string

	if applyResult.Success {
		color = teamsSuccessColor
		successWord = "succeeded"
		activityImage = "https://raw.githubusercontent.com/runatlantis/atlantis/main/runatlantis.io/public/hero.png"
	} else {
		color = teamsFailureColor
		successWord = "failed"
		activityImage = "https://raw.githubusercontent.com/runatlantis/atlantis/main/runatlantis.io/public/hero.png"
	}

	title := fmt.Sprintf("Atlantis Apply %s", successWord)
	subtitle := fmt.Sprintf("Repository: %s", applyResult.Repo.FullName)
	summary := fmt.Sprintf("Terraform apply %s for %s", successWord, applyResult.Repo.FullName)

	directory := applyResult.Directory
	// Since "." looks weird, replace it with "/" to make it clear this is the root.
	if directory == "." {
		directory = "/"
	}

	facts := []TeamsMessageFact{
		{
			Name:  "Workspace",
			Value: applyResult.Workspace,
		},
		{
			Name:  "Branch",
			Value: applyResult.Pull.BaseBranch,
		},
		{
			Name:  "User",
			Value: applyResult.User.Username,
		},
		{
			Name:  "Directory",
			Value: directory,
		},
		{
			Name:  "Pull Request",
			Value: fmt.Sprintf("[#%d](%s)", applyResult.Pull.Num, applyResult.Pull.URL),
		},
	}

	if applyResult.ProjectName != "" {
		facts = append(facts, TeamsMessageFact{
			Name:  "Project",
			Value: applyResult.ProjectName,
		})
	}

	return TeamsMessage{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: color,
		Summary:    summary,
		Sections: []TeamsMessageSection{
			{
				ActivityTitle:    title,
				ActivitySubtitle: subtitle,
				ActivityImage:    activityImage,
				Facts:            facts,
				Markdown:         true,
			},
		},
	}
}
