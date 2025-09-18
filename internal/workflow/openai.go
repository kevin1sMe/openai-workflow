package workflow

import (
	"fmt"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type ClientOptions struct {
	APIKey  string
	OrgID   string
	BaseURL string
}

func NewClient(opts ClientOptions) (*openai.Client, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("openai_api_key not set")
	}
	var clientOpts []option.RequestOption
	clientOpts = append(clientOpts, option.WithAPIKey(opts.APIKey))
	if opts.OrgID != "" {
		clientOpts = append(clientOpts, option.WithOrganization(opts.OrgID))
	}
	if opts.BaseURL != "" {
		clientOpts = append(clientOpts, option.WithBaseURL(opts.BaseURL))
	}
	client := openai.NewClient(clientOpts...)
	return &client, nil
}
