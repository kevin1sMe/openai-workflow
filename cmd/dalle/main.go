package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	openai "github.com/openai/openai-go"

	"github.com/openai-workflow/workflow/internal/workflow"
)

type alfredResponse struct {
	Response  string            `json:"response,omitempty"`
	Rerun     float64           `json:"rerun,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
	Behaviour map[string]string `json:"behaviour,omitempty"`
	Footer    string            `json:"footer,omitempty"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	env, err := workflow.LoadEnv()
	if err != nil {
		return respondError(err)
	}
	if env.APIKey == "" {
		return respondError(fmt.Errorf("OpenAI API key missing"))
	}

	dalleEnv, err := workflow.LoadDalleEnv()
	if err != nil {
		return respondError(err)
	}

	typedQuery := ""
	if len(os.Args) > 1 {
		typedQuery = os.Args[1]
	}

	previous, err := previousMarkdown(dalleEnv.ParentFolder, 10)
	if err != nil {
		return respondError(err)
	}

	previousResponse := strings.Join(previous, "\n\n")
	if previousResponse == "" {
		previousResponse = ""
	}

	variables := map[string]string{"loaded_previous": "true"}

	if os.Getenv("loaded_previous") == "" {
		resp := alfredResponse{
			Rerun:     0.1,
			Variables: variables,
			Response:  previousResponse,
			Behaviour: map[string]string{"scroll": "end"},
		}
		return emit(resp)
	}

	if typedQuery == "" {
		resp := alfredResponse{
			Response:  previousResponse,
			Variables: variables,
			Behaviour: map[string]string{"scroll": "end"},
		}
		return emit(resp)
	}

	client, err := workflow.NewClient(workflow.ClientOptions{
		APIKey:  env.APIKey,
		OrgID:   env.OrgID,
		BaseURL: workflow.NormalizeBaseURL(env.DalleAPIEndpoint, "https://api.openai.com/v1", "/images/generations"),
	})
	if err != nil {
		return respondError(err)
	}

	params := openai.ImageGenerateParams{
		Prompt: typedQuery,
	}
	params.N = openai.Int(1)
	params.ResponseFormat = openai.ImageGenerateParamsResponseFormatURL
	params.Size = openai.ImageGenerateParamsSize("1024x1024")

	if dalleEnv.Model != "" {
		params.Model = openai.ImageModel(dalleEnv.Model)
	}
	if dalleEnv.Style != "" {
		params.Style = openai.ImageGenerateParamsStyle(dalleEnv.Style)
	}
	if dalleEnv.Quality != "" {
		params.Quality = openai.ImageGenerateParamsQuality(dalleEnv.Quality)
	}

	ctx := context.Background()
	imagesResp, err := client.Images.Generate(ctx, params)
	if err != nil {
		return respondWithPreviousError(previousResponse, typedQuery, err)
	}

	creation := time.Unix(imagesResp.Created, 0)
	downloaded, err := downloadImages(ctx, imagesResp.Data, typedQuery, dalleEnv, creation)
	if err != nil {
		return respondWithPreviousError(previousResponse, typedQuery, err)
	}

	markdown := make([]string, 0, len(downloaded))
	for _, path := range downloaded {
		markdown = append(markdown, workflow.MarkdownImage(path))
	}

	resp := alfredResponse{
		Response:  strings.Join(markdown, "\n\n"),
		Variables: variables,
		Behaviour: map[string]string{"response": "append"},
	}
	return emit(resp)
}

func previousMarkdown(folder string, max int) ([]string, error) {
	images, err := workflow.LatestImages(folder, max)
	if err != nil {
		return nil, err
	}
	var markdown []string
	for _, img := range images {
		markdown = append(markdown, workflow.MarkdownImage(img))
	}
	return markdown, nil
}

func downloadImages(ctx context.Context, data []openai.Image, prompt string, dalleEnv *workflow.DalleEnv, creation time.Time) ([]string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	var paths []string

	for _, item := range data {
		if item.URL == "" {
			return nil, fmt.Errorf("image response missing URL")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("image download failed: %s", resp.Status)
		}

		uid := workflow.RandomUID()
		filename := workflow.BuildImageFilename(dalleEnv.ParentFolder, creation, uid)
		if err := os.WriteFile(filename, body, 0o644); err != nil {
			return nil, err
		}

		if dalleEnv.IncludeMetadata {
			promptText := buildPromptText(prompt, item.RevisedPrompt)
			description := fmt.Sprintf("Generated with DALL-E via Alfred workflow.\n\n%s", promptText)
			if err := workflow.WriteMetadata("kMDItemCreator", "DALL-E", filename); err != nil {
				fmt.Fprintln(os.Stderr, "metadata error:", err)
			}
			if err := workflow.WriteMetadata("kMDItemDescription", description, filename); err != nil {
				fmt.Fprintln(os.Stderr, "metadata error:", err)
			}
		}

		paths = append(paths, filename)
	}

	return paths, nil
}

func buildPromptText(original, revised string) string {
	if strings.TrimSpace(revised) == "" {
		return fmt.Sprintf("Original Prompt: %s", original)
	}
	return fmt.Sprintf("Original Prompt: %s\n\nRevised Prompt: %s", original, revised)
}

func respondWithPreviousError(previous, prompt string, err error) error {
	message := previous
	if message != "" {
		message += "\n\n"
	}
	message += fmt.Sprintf("**Original Prompt:** %s\n\n%s", prompt, err.Error())
	resp := alfredResponse{
		Response:  message,
		Behaviour: map[string]string{"response": "append", "scroll": "end"},
	}
	return emit(resp)
}

func respondError(err error) error {
	resp := alfredResponse{Response: err.Error()}
	return emit(resp)
}

func emit(resp alfredResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
