package eino

import (
	"context"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
)

type fakeChatModelFactory struct {
	model einomodel.BaseChatModel
	err   error
}

func (f fakeChatModelFactory) Model(context.Context) (einomodel.BaseChatModel, error) {
	return f.model, f.err
}

func (f fakeChatModelFactory) ModelMetadata() map[string]any {
	return map[string]any{"provider": "fake", "model": "fake"}
}

type fakeChatModel struct {
	content string
	err     error
}

func (m fakeChatModel) Generate(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &schema.Message{
		Role:    schema.Assistant,
		Content: m.content,
	}, nil
}

func (m fakeChatModel) Stream(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func TestRunModelOnceReturnsText(t *testing.T) {
	out, err := runModelOnce(t.Context(), fakeChatModelFactory{
		model: fakeChatModel{content: `{"summary":"ok"}`},
	}, "test", "instruction", "input")

	require.NoError(t, err)
	require.Equal(t, `{"summary":"ok"}`, out.Text)
}

func TestRunModelOnceRequiresContent(t *testing.T) {
	_, err := runModelOnce(context.Background(), fakeChatModelFactory{
		model: fakeChatModel{content: "  "},
	}, "test", "instruction", "input")

	require.ErrorContains(t, err, "empty response")
}

func TestChatModelFactoryConfigValidation(t *testing.T) {
	disabled := newChatModelProvider(rez.AiConfig{Enabled: false})
	_, disabledErr := disabled.Model(context.Background())
	require.ErrorContains(t, disabledErr, "disabled")

	missingKey := newChatModelProvider(rez.AiConfig{
		Enabled:  true,
		Provider: "gemini",
		Model:    "gemini-2.5-flash",
	})
	_, missingKeyErr := missingKey.Model(context.Background())
	require.ErrorContains(t, missingKeyErr, "api key")
}
