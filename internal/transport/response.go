package transport

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
)

func decodeSuccess(body []byte, output any) error {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		return json.Unmarshal(envelope.Data, output)
	}
	return json.Unmarshal(body, output)
}

func decodeAPIError(status int, requestID, key string, body []byte) error {
	result := &api.APIError{StatusCode: status, RequestID: requestID, IdempotencyKey: key, Message: http.StatusText(status)}
	var envelope struct {
		Code      string          `json:"code"`
		Message   string          `json:"message"`
		Msg       string          `json:"msg"`
		RequestID string          `json:"request_id"`
		Details   json.RawMessage `json:"details"`
		Error     *struct {
			Code      string          `json:"code"`
			Message   string          `json:"message"`
			RequestID string          `json:"request_id"`
			Details   json.RawMessage `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil {
		result.Code = envelope.Code
		result.Message = firstNonEmpty(envelope.Message, envelope.Msg, result.Message)
		result.RequestID = firstNonEmpty(result.RequestID, envelope.RequestID)
		result.Details = envelope.Details
		if envelope.Error != nil {
			result.Code = envelope.Error.Code
			result.Message = firstNonEmpty(envelope.Error.Message, result.Message)
			result.RequestID = firstNonEmpty(result.RequestID, envelope.Error.RequestID)
			result.Details = envelope.Error.Details
		}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
