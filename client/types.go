package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageRole is the role of a conversation message.
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
	RoleDeveloper MessageRole = "developer"
)

// MessagesResponse is the response body for GET /api/actors/{actor_id}/messages.
type MessagesResponse struct {
	Messages []ConversationMessage `json:"messages"`
}

// ConversationMessage is a single message in an actor conversation.
type ConversationMessage struct {
	ID          string          `json:"id"`
	Role        MessageRole     `json:"role"`
	Content     *MessageContent `json:"content"`
	Name        *string         `json:"name"`
	ToolCallID  *string         `json:"tool_call_id"`
	Refusal     *string         `json:"refusal"`
	Position    int             `json:"position"`
	InsertedAt  time.Time       `json:"inserted_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// MessageContent is message body content (text or parts).
type MessageContent struct {
	Text  *TextContent
	Parts *PartsContent
}

// TextContent is plain text message content.
type TextContent struct {
	Type  string `json:"type"` // always "text"
	Value string `json:"value"`
}

// PartsContent is multi-part message content.
type PartsContent struct {
	Type  string        `json:"type"` // always "parts"
	Value []ContentPart `json:"value"`
}

// ContentPart is a single segment inside parts content.
// Additional vendor-specific fields are stored in Extra.
type ContentPart struct {
	Type       string          `json:"type"`
	Text       *string         `json:"text,omitempty"`
	Thinking   *string         `json:"thinking,omitempty"`
	Signature  *string         `json:"signature,omitempty"`
	Source     json.RawMessage `json:"source,omitempty"`
	ID         *string         `json:"id,omitempty"`
	Name       *string         `json:"name,omitempty"`
	Arguments  json.RawMessage `json:"arguments,omitempty"`
	ToolCallID *string         `json:"tool_call_id,omitempty"`
	Content    json.RawMessage `json:"content,omitempty"`
	IsError    *bool           `json:"is_error,omitempty"`
	Extra      map[string]any  `json:"-"`
}

// SuccessResponse is returned when an operation succeeds with no payload.
type SuccessResponse struct {
	Success bool `json:"success"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail holds a human-readable error message.
type ErrorDetail struct {
	Message string `json:"message"`
}

// ToolsUpdateBody is a flat map of tool name → definition (PUT /tools body variant).
type ToolsUpdateBody map[string]ToolDefinition

// ToolsUpdateWrapped wraps tools in a "tools" field (PUT /tools body variant).
type ToolsUpdateWrapped struct {
	Tools map[string]ToolDefinition `json:"tools"`
}

// ToolDefinition describes a single callable tool for the actor.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Parameters  *JsonSchema     `json:"parameters,omitempty"`
	InputSchema *JsonSchema     `json:"input_schema,omitempty"`
	Strict      *bool           `json:"strict,omitempty"`
	Extra       map[string]any  `json:"-"`
}

// JsonSchema is a JSON Schema subset used for tool parameters.
type JsonSchema struct {
	Type                 *string                `json:"type,omitempty"`
	Description          *string                `json:"description,omitempty"`
	Properties           map[string]*JsonSchema `json:"properties,omitempty"`
	Items                json.RawMessage        `json:"items,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Enum                 []json.RawMessage      `json:"enum,omitempty"`
	AdditionalProperties *bool                  `json:"additionalProperties,omitempty"`
	Extra                map[string]any         `json:"-"`
}

func (c *MessageContent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	switch probe.Type {
	case "text":
		var t TextContent
		if err := json.Unmarshal(data, &t); err != nil {
			return err
		}
		c.Text = &t
	case "parts":
		var p PartsContent
		if err := json.Unmarshal(data, &p); err != nil {
			return err
		}
		c.Parts = &p
	default:
		return fmt.Errorf("client: unknown message content type %q", probe.Type)
	}
	return nil
}

func (c MessageContent) MarshalJSON() ([]byte, error) {
	if c.Text != nil {
		return json.Marshal(c.Text)
	}
	if c.Parts != nil {
		return json.Marshal(c.Parts)
	}
	return []byte("null"), nil
}

func (p *ContentPart) UnmarshalJSON(data []byte) error {
	type contentPart ContentPart
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	known := map[string]struct{}{
		"type": {}, "text": {}, "thinking": {}, "signature": {}, "source": {},
		"id": {}, "name": {}, "arguments": {}, "tool_call_id": {}, "content": {},
		"is_error": {},
	}
	extra := make(map[string]any)
	for k, v := range raw {
		if _, ok := known[k]; !ok {
			var anyVal any
			if err := json.Unmarshal(v, &anyVal); err != nil {
				return err
			}
			extra[k] = anyVal
		}
	}
	var base contentPart
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	*p = ContentPart(base)
	if len(extra) > 0 {
		p.Extra = extra
	}
	return nil
}

func (p ContentPart) MarshalJSON() ([]byte, error) {
	type contentPart ContentPart
	out, err := json.Marshal(contentPart(p))
	if err != nil {
		return nil, err
	}
	if len(p.Extra) == 0 {
		return out, nil
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		return nil, err
	}
	for k, v := range p.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

func (t *ToolDefinition) UnmarshalJSON(data []byte) error {
	type toolDef ToolDefinition
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	known := map[string]struct{}{
		"name": {}, "description": {}, "parameters": {},
		"input_schema": {}, "strict": {},
	}
	extra := make(map[string]any)
	for k, v := range raw {
		if _, ok := known[k]; !ok {
			var anyVal any
			if err := json.Unmarshal(v, &anyVal); err != nil {
				return err
			}
			extra[k] = anyVal
		}
	}
	var base toolDef
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	*t = ToolDefinition(base)
	if len(extra) > 0 {
		t.Extra = extra
	}
	return nil
}

func (t ToolDefinition) MarshalJSON() ([]byte, error) {
	type toolDef ToolDefinition
	out, err := json.Marshal(toolDef(t))
	if err != nil {
		return nil, err
	}
	if len(t.Extra) == 0 {
		return out, nil
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		return nil, err
	}
	for k, v := range t.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

func (s *JsonSchema) UnmarshalJSON(data []byte) error {
	type jsonSchema JsonSchema
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	known := map[string]struct{}{
		"type": {}, "description": {}, "properties": {}, "items": {},
		"required": {}, "enum": {}, "additionalProperties": {},
	}
	extra := make(map[string]any)
	for k, v := range raw {
		if _, ok := known[k]; !ok {
			var anyVal any
			if err := json.Unmarshal(v, &anyVal); err != nil {
				return err
			}
			extra[k] = anyVal
		}
	}
	var base jsonSchema
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	*s = JsonSchema(base)
	if len(extra) > 0 {
		s.Extra = extra
	}
	return nil
}

func (s JsonSchema) MarshalJSON() ([]byte, error) {
	type jsonSchema JsonSchema
	out, err := json.Marshal(jsonSchema(s))
	if err != nil {
		return nil, err
	}
	if len(s.Extra) == 0 {
		return out, nil
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		return nil, err
	}
	for k, v := range s.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}
