package local

type evaluationVariant struct {
	Key     string      `json:"key,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type flagResult struct {
	Variant          evaluationVariant `json:"variant,omitempty"`
	Description      string            `json:"description,omitempty"`
	IsDefaultVariant bool              `json:"isDefaultVariant,omitempty"`
}

type EvaluationResult = map[string]flagResult

type interopResult struct {
	Result *EvaluationResult `json:"result,omitempty"`
	Error  *string           `json:"error,omitempty"`
}
