// File: internal/destination/prompt.go
package destination

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

type Prompt struct{}

func NewPrompt() *Prompt {
	return &Prompt{}
}

func (p *Prompt) PromptDestination(defaultDest *Destination) (string, Destination, error) {
	// Destination name
	namePrompt := promptui.Prompt{
		Label:     "Destination Name",
		AllowEdit: true,
		Validate:  validateNotEmpty,
	}
	name, err := namePrompt.Run()
	if err != nil {
		return "", Destination{}, fmt.Errorf("destination name prompt failed: %w", err)
	}

	// Destination type
	typePrompt := promptui.Select{
		Label: "Destination Type",
		Items: []string{"slack", "lineworks", "custom"},
	}
	_, destType, err := typePrompt.Run()
	if err != nil {
		return "", Destination{}, fmt.Errorf("destination type prompt failed: %w", err)
	}

	var dest Destination
	dest.Type = destType

	// Get the rest of the configuration based on type
	switch destType {
	case "slack":
		if err := p.promptSlackConfig(&dest, defaultDest); err != nil {
			return "", Destination{}, err
		}

	case "lineworks":
		if err := p.promptLineworksConfig(&dest, defaultDest); err != nil {
			return "", Destination{}, err
		}

	case "custom":
		if err := p.promptCustomConfig(&dest, defaultDest); err != nil {
			return "", Destination{}, err
		}
	}

	return name, dest, nil
}

func (p *Prompt) promptSlackConfig(dest *Destination, defaultDest *Destination) error {
	// Get default values
	defaultToken := ""
	defaultChannel := ""
	if defaultDest != nil && defaultDest.Type == "slack" {
		defaultToken = defaultDest.Token.Value
		defaultChannel = defaultDest.Channel
	}

	// Token
	tokenPrompt := promptui.Prompt{
		Label:     "Slack Bot Token (starts with xoxb-)",
		Validate:  validateSlackToken,
		Mask:      '*',
		AllowEdit: true,
		Default:   defaultToken,
	}
	token, err := tokenPrompt.Run()
	if err != nil {
		return fmt.Errorf("token prompt failed: %w", err)
	}
	dest.Token = TokenConfig{
		Type:  "bot",
		Value: token,
	}

	// Channel
	channelPrompt := promptui.Prompt{
		Label:     "Slack Channel",
		Validate:  validateSlackChannel,
		AllowEdit: true,
		Default:   defaultChannel,
	}
	channel, err := channelPrompt.Run()
	if err != nil {
		return fmt.Errorf("channel prompt failed: %w", err)
	}
	dest.Channel = channel

	return nil
}

func (p *Prompt) promptLineworksConfig(dest *Destination, defaultDest *Destination) error {
	// Get default values
	defaultURL := ""
	defaultChannel := ""
	if defaultDest != nil && defaultDest.Type == "lineworks" {
		defaultURL = defaultDest.URL
		defaultChannel = defaultDest.Channel
	}

	// URL
	urlPrompt := promptui.Prompt{
		Label:     "Lineworks Webhook URL",
		Validate:  validateURL,
		AllowEdit: true,
		Default:   defaultURL,
	}
	url, err := urlPrompt.Run()
	if err != nil {
		return fmt.Errorf("URL prompt failed: %w", err)
	}
	dest.URL = url

	// Channel
	channelPrompt := promptui.Prompt{
		Label:     "Lineworks Channel",
		Validate:  validateNotEmpty,
		AllowEdit: true,
		Default:   defaultChannel,
	}
	channel, err := channelPrompt.Run()
	if err != nil {
		return fmt.Errorf("channel prompt failed: %w", err)
	}
	dest.Channel = channel

	return nil
}

func (p *Prompt) promptCustomConfig(dest *Destination, defaultDest *Destination) error {
	// Get default values
	defaultURL := ""
	defaultTokenValue := ""
	if defaultDest != nil && defaultDest.Type == "custom" {
		defaultURL = defaultDest.URL
		defaultTokenValue = defaultDest.Token.Value
	}

	// URL
	urlPrompt := promptui.Prompt{
		Label:     "API URL",
		Validate:  validateURL,
		AllowEdit: true,
		Default:   defaultURL,
	}
	url, err := urlPrompt.Run()
	if err != nil {
		return fmt.Errorf("URL prompt failed: %w", err)
	}
	dest.URL = url

	// Token type
	tokenTypePrompt := promptui.Select{
		Label: "Authentication Type",
		Items: []string{"bearer", "basic", "api_key", "none"},
	}
	_, tokenType, err := tokenTypePrompt.Run()
	if err != nil {
		return fmt.Errorf("auth type prompt failed: %w", err)
	}

	if tokenType != "none" {
		// Token value
		tokenPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("%s Token", tokenType),
			Validate:  validateNotEmpty,
			Mask:      '*',
			AllowEdit: true,
			Default:   defaultTokenValue,
		}
		tokenValue, err := tokenPrompt.Run()
		if err != nil {
			return fmt.Errorf("token prompt failed: %w", err)
		}
		dest.Token = TokenConfig{
			Type:  tokenType,
			Value: tokenValue,
		}
	}

	return nil
}

// Validation functions
func validateNotEmpty(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func validateURL(input string) error {
	if err := validateNotEmpty(input); err != nil {
		return err
	}
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	return nil
}

func validateSlackToken(input string) error {
	if err := validateNotEmpty(input); err != nil {
		return err
	}
	if !strings.HasPrefix(input, "xoxb-") {
		return fmt.Errorf("slack bot token must start with 'xoxb-'")
	}
	return nil
}

func validateSlackChannel(input string) error {
	if err := validateNotEmpty(input); err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "#") && !strings.HasPrefix(input, "C") {
		return fmt.Errorf("slack channel must start with '#' or be a channel ID")
	}
	return nil
}
