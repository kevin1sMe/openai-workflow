package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
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

const (
	streamModeEnv = "GOCHAT_MODE"
	streamModeRun = "stream"
)

func main() {
	if len(os.Args) > 2 && os.Args[1] == "--dump-chat" {
		if err := dumpChat(os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if os.Getenv(streamModeEnv) == streamModeRun {
		if err := runStreamProcess(); err != nil {
			fmt.Fprintln(os.Stderr, "stream error:", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--stream" {
		if err := runStreamProcess(); err != nil {
			fmt.Fprintln(os.Stderr, "stream error:", err)
			os.Exit(1)
		}
		return
	}

	if err := runPrimary(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runPrimary() error {
	env, err := workflow.LoadEnv()
	if err != nil {
		return respondError(err)
	}
	if err := workflow.EnsureHelperBinary(env.WorkflowDataDir); err != nil {
		return respondError(err)
	}
	if env.APIKey == "" {
		return respondError(errors.New("OpenAI API key missing"))
	}

	if err := workflow.EnsureChatFile(env.ChatFile); err != nil {
		return respondError(err)
	}

	args := os.Args[1:]
	typedQuery := ""
	if len(args) > 0 {
		typedQuery = args[0]
	}

	streamingNow := os.Getenv("streaming_now") == "1"
	streamMarker := os.Getenv("stream_marker") == "1"

	if streamingNow {
		return respondStream(env, streamMarker)
	}

	// Resume stream if files linger
	if workflow.StreamFileExists(env.StreamFile) {
		resp := alfredResponse{
			Rerun: 0.1,
			Variables: map[string]string{
				"streaming_now": "1",
				"stream_marker": "1",
			},
		}
		chat, err := workflow.ReadChat(env.ChatFile)
		if err == nil {
			resp.Response = workflow.MarkdownChat(chat, true)
			resp.Behaviour = map[string]string{"scroll": "end"}
		}
		return emit(resp)
	}

	chat, err := workflow.ReadChat(env.ChatFile)
	if err != nil {
		return respondError(err)
	}

	if typedQuery == "" {
		resp := alfredResponse{
			Response:  workflow.MarkdownChat(chat, false),
			Behaviour: map[string]string{"scroll": "end"},
		}
		return emit(resp)
	}

	appendMsg := workflow.Message{Role: "user", Content: typedQuery}
	chat = append(chat, appendMsg)
	if err := workflow.WriteChat(env.ChatFile, chat); err != nil {
		return respondError(err)
	}

	if err := workflow.Touch(env.StreamFile); err != nil {
		return respondError(err)
	}

	if err := startBackgroundStream(env); err != nil {
		return respondError(err)
	}

	resp := alfredResponse{
		Rerun: 0.1,
		Variables: map[string]string{
			"streaming_now": "1",
			"stream_marker": "1",
		},
		Response: workflow.MarkdownChat(chat, true),
	}
	return emit(resp)
}

func startBackgroundStream(env *workflow.Env) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(executable, "--stream")
	cmd.Env = append(os.Environ(), streamModeEnv+"="+streamModeRun)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	pid := fmt.Sprintf("%d", cmd.Process.Pid)
	if err := os.WriteFile(env.PIDFile, []byte(pid), 0o644); err != nil {
		return err
	}
	return nil
}

func runStreamProcess() error {
	env, err := workflow.LoadEnv()
	if err != nil {
		return err
	}
	if err := workflow.EnsureHelperBinary(env.WorkflowDataDir); err != nil {
		return err
	}
	if env.APIKey == "" {
		return workflow.WriteStreamState(env.StreamFile, workflow.StreamState{Error: "Missing OpenAI API key"})
	}

	if err := runChatStream(env); err != nil {
		workflow.WriteStreamState(env.StreamFile, workflow.StreamState{Error: err.Error()})
		return err
	}
	return nil
}

func runChatStream(env *workflow.Env) error {
	client, err := workflow.NewClient(workflow.ClientOptions{
		APIKey:  env.APIKey,
		OrgID:   env.OrgID,
		BaseURL: workflow.NormalizeBaseURL(env.ChatAPIEndpoint, "https://api.openai.com/v1", "/chat/completions"),
	})
	if err != nil {
		return err
	}

	if err := workflow.WriteStreamState(env.StreamFile, workflow.StreamState{}); err != nil {
		return err
	}

	chat, err := workflow.ReadChat(env.ChatFile)
	if err != nil {
		return err
	}

	trimmed := workflow.TrimContext(chat, env.MaxContext)
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(trimmed)+1)
	if env.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(env.SystemPrompt))
	}
	for _, m := range trimmed {
		switch m.Role {
		case "user":
			messages = append(messages, openai.UserMessage(m.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(m.Content))
		}
	}

	model := workflow.ResolveChatModel(env.GPTModel, env.ChatModelOverride)
	if model == "" {
		return errors.New("gpt_model not configured")
	}

	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:    model,
		Messages: messages,
	})

	acc := openai.ChatCompletionAccumulator{}
	builder := strings.Builder{}

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				builder.WriteString(delta)
				workflow.WriteStreamState(env.StreamFile, workflow.StreamState{Content: builder.String()})
			}
		}
	}

	if err := stream.Err(); err != nil {
		workflow.WriteStreamState(env.StreamFile, workflow.StreamState{Error: err.Error(), Content: builder.String()})
		return err
	}

	finishReason := ""
	if len(acc.Choices) > 0 {
		finishReason = acc.Choices[0].FinishReason
	}

	return workflow.WriteStreamState(env.StreamFile, workflow.StreamState{
		Content:      builder.String(),
		FinishReason: finishReason,
	})
}

func respondStream(env *workflow.Env, marker bool) error {
	if marker {
		resp := alfredResponse{
			Rerun:     0.1,
			Variables: map[string]string{"streaming_now": "1"},
			Response:  "…",
			Behaviour: map[string]string{"response": "append"},
		}
		return emit(resp)
	}

	state, err := workflow.ReadStreamState(env.StreamFile)
	if err != nil {
		return respondError(err)
	}

	if state.Error != "" {
		workflow.RemoveFiles(env.StreamFile, env.PIDFile)
		resp := alfredResponse{
			Response:  state.Error,
			Behaviour: map[string]string{"response": "replacelast"},
		}
		return emit(resp)
	}

	now := time.Now()
	age, err := workflow.FileAge(env.StreamFile, now)
	stalled := err == nil && age > time.Duration(env.TimeoutSeconds)*time.Second

	if state.FinishReason == "" && !stalled {
		resp := alfredResponse{
			Rerun:     0.1,
			Variables: map[string]string{"streaming_now": "1"},
			Response:  state.Content,
			Behaviour: map[string]string{"response": "replacelast", "scroll": "end"},
		}
		return emit(resp)
	}

	chat, err := workflow.ReadChat(env.ChatFile)
	if err != nil {
		return respondError(err)
	}

	assistantMessage := workflow.Message{Role: "assistant", Content: state.Content}
	if state.Content != "" {
		chat = append(chat, assistantMessage)
		if err := workflow.WriteChat(env.ChatFile, chat); err != nil {
			return respondError(err)
		}
	}

	workflow.RemoveFiles(env.StreamFile, env.PIDFile)

	footer := footerForFinish(state.FinishReason)
	if stalled {
		footer = "You can ask ChatGPT to continue the answer"
	}

	responseText := state.Content
	if stalled {
		responseText = strings.TrimSpace(state.Content) + " [Connection Stalled]"
	}

	resp := alfredResponse{
		Response:  responseText,
		Footer:    footer,
		Behaviour: map[string]string{"response": "replacelast", "scroll": "end"},
	}
	return emit(resp)
}

func footerForFinish(reason string) string {
	switch reason {
	case "length":
		return "Maximum number of tokens reached"
	case "content_filter":
		return "Content was omitted due to a flag from OpenAI content filters"
	default:
		return ""
	}
}

func dumpChat(path string) error {
	messages, err := workflow.ReadChat(path)
	if err != nil {
		return err
	}
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func respondError(err error) error {
	resp := alfredResponse{
		Response: err.Error(),
	}
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
