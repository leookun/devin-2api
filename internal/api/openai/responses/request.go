// 本文件定义 OpenAI Responses 请求 JSON 到中间 LLM 模型的转换。
//
// Package responses 定义 OpenAI Responses HTTP 协议与中间 LLM 模型之间的编解码。
package responses

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/leookun/devin-2api/internal/api/common"
	"github.com/leookun/devin-2api/internal/llm"
)

// Request 是 OpenAI Responses 请求中本适配器支持的字段集合。
type Request struct {
	// Model 是请求使用的模型标识。
	Model string `json:"model"`
	// Instructions 是独立于 input 的系统提示词。
	Instructions string `json:"instructions,omitempty"`
	// Input 是字符串或 Responses input item 数组。
	Input json.RawMessage `json:"input"`
	// Tools 是 OpenAI function 工具定义。
	Tools []Tool `json:"tools,omitempty"`
	// Stream 表示是否请求流式响应。
	Stream bool `json:"stream,omitempty"`
	// MaxOutputTokens 是可选的输出 token 上限。
	MaxOutputTokens *int `json:"max_output_tokens,omitempty"`
	// Temperature 是可选的采样温度。
	Temperature *float64 `json:"temperature,omitempty"`
	// PreviousResponseID 是上游 Responses 会话关联标识。
	PreviousResponseID string `json:"previous_response_id,omitempty"`
}

// Tool 是 OpenAI Responses function 工具定义。
type Tool struct {
	// Type 固定为 function。
	Type string `json:"type"`
	// Name 是工具名称。
	Name string `json:"name"`
	// Description 是工具用途说明。
	Description string `json:"description,omitempty"`
	// Parameters 是工具输入 JSON Schema。
	Parameters json.RawMessage `json:"parameters"`
}

// AdaptedRequest 是 OpenAI 请求转换后的中间请求和生成选项。
type AdaptedRequest struct {
	// Context 是供应商无关的完整对话上下文。
	Context llm.RequestMessages
	// Options 是本次生成所需的协议选项。
	Options RequestOptions
}

// RequestOptions 保存不属于对话历史的生成控制参数。
type RequestOptions struct {
	// Stream 表示调用方是否请求流式响应。
	Stream bool
	// MaxOutputTokens 是可选的输出 token 上限。
	MaxOutputTokens *int
	// Temperature 是可选的采样温度。
	Temperature *float64
	// PreviousResponseID 是调用方提供的上游响应关联标识。
	PreviousResponseID string
}

// DecodeRequest 将 OpenAI Responses JSON 请求转换为中间请求。
func DecodeRequest(data []byte) (AdaptedRequest, error) {
	var request Request
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&request); err != nil {
		return AdaptedRequest{}, fmt.Errorf("decode responses request: %w", err)
	}
	if request.Model == "" {
		return AdaptedRequest{}, errors.New("responses request model is required")
	}

	context := llm.RequestMessages{Model: request.Model, SystemPrompt: request.Instructions}
	if err := appendInputMessages(&context, request.Input); err != nil {
		return AdaptedRequest{}, err
	}
	for _, tool := range request.Tools {
		if tool.Type != "function" {
			continue
		}
		schema := tool.Parameters
		if len(schema) == 0 {
			schema = json.RawMessage(`{"type":"object"}`)
		}
		context.Tools = append(context.Tools, llm.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: schema,
		})
	}
	if err := context.Validate(); err != nil {
		return AdaptedRequest{}, fmt.Errorf("validate adapted request: %w", err)
	}
	return AdaptedRequest{
		Context: context,
		Options: RequestOptions{
			Stream:             request.Stream,
			MaxOutputTokens:    request.MaxOutputTokens,
			Temperature:        request.Temperature,
			PreviousResponseID: request.PreviousResponseID,
		},
	}, nil
}

func appendInputMessages(context *llm.RequestMessages, raw json.RawMessage) error {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		context.Messages = append(context.Messages, llm.UserMessage{
			Content:     []llm.Content{llm.TextContent{Text: text}},
			TimestampMS: time.Now().UnixMilli(),
		})
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("decode responses input: %w", err)
	}
	for index, item := range items {
		if err := appendInputItem(context, item); err != nil {
			return fmt.Errorf("input[%d]: %w", index, err)
		}
	}
	return nil
}

func appendInputItem(context *llm.RequestMessages, raw json.RawMessage) error {
	var header struct {
		Type string `json:"type"`
		Role string `json:"role"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return fmt.Errorf("decode input item: %w", err)
	}
	if header.Type == "" && header.Role != "" {
		header.Type = "message"
	}
	switch header.Type {
	case "message":
		return appendMessageItem(context, raw, header.Role)
	case "function_call":
		var item struct {
			CallID    string `json:"call_id"`
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return err
		}
		arguments := json.RawMessage(item.Arguments)
		call := llm.ToolCall{ID: item.CallID, Name: item.Name, Arguments: arguments}
		// Devin 要求同一次生成发出的多个工具调用落在同一条 assistant 消息里：
		// “两条 assistant 各带一个 toolCall，后面跟两个结果” 会被上游判为非法序列并
		// 回 invalid_argument（github.com/leookun/devin-2api/issues/7）。Responses
		// 协议把每次调用拆成独立 item，这里并回上一条仍是 tool-use 的消息。
		if last := len(context.Messages) - 1; last >= 0 {
			if assistant, ok := context.Messages[last].(llm.AssistantMessage); ok &&
				assistant.StopReason == llm.StopReasonToolUse {
				assistant.Content = append(assistant.Content, call)
				context.Messages[last] = assistant
				return nil
			}
		}
		context.Messages = append(context.Messages, llm.AssistantMessage{
			Content:     []llm.Content{call},
			StopReason:  llm.StopReasonToolUse,
			TimestampMS: time.Now().UnixMilli(),
		})
		return nil
	case "function_call_output":
		var item struct {
			CallID string          `json:"call_id"`
			Output json.RawMessage `json:"output"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return err
		}
		output, err := common.RawOutputText(item.Output)
		if err != nil {
			return err
		}
		toolName := findToolName(context.Messages, item.CallID)
		if toolName == "" {
			return fmt.Errorf("function call output references unknown call_id %q", item.CallID)
		}
		context.Messages = append(context.Messages, llm.ToolResultMessage{
			ToolCallID:  item.CallID,
			ToolName:    toolName,
			Content:     []llm.Content{llm.TextContent{Text: output}},
			TimestampMS: time.Now().UnixMilli(),
		})
		return nil
	default:
		return nil
	}
}

func findToolName(messages []llm.Message, callID string) string {
	for index := len(messages) - 1; index >= 0; index-- {
		assistant, ok := messages[index].(llm.AssistantMessage)
		if !ok {
			continue
		}
		for _, block := range assistant.Content {
			call, ok := block.(llm.ToolCall)
			if ok && call.ID == callID {
				return call.Name
			}
		}
	}
	return ""
}

func appendMessageItem(context *llm.RequestMessages, raw json.RawMessage, role string) error {
	switch role {
	case "user", "assistant", "system", "developer":
	default:
		return nil
	}
	var item struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return err
	}
	content, err := common.DecodeContent(item.Content)
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return nil
	}
	switch role {
	case "user":
		context.Messages = append(context.Messages, llm.UserMessage{Content: content, TimestampMS: time.Now().UnixMilli()})
	case "assistant":
		context.Messages = append(context.Messages, llm.AssistantMessage{Content: content, TimestampMS: time.Now().UnixMilli()})
	case "system", "developer":
		text := common.ContentText(content)
		if context.SystemPrompt != "" && text != "" {
			context.SystemPrompt += "\n"
		}
		context.SystemPrompt += text
	}
	return nil
}
