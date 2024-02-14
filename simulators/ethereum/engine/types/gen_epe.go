// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/core-coin/go-core/common/hexutil"
)

var _ = (*executionPayloadEnvelopeMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (e ExecutionPayloadEnvelope) MarshalJSON() ([]byte, error) {
	type ExecutionPayloadEnvelope struct {
		ExecutionPayload      *ExecutableData `json:"executionPayload"       gencodec:"required"`
		BlockValue            *hexutil.Big    `json:"blockValue"             gencodec:"required"`
		BlobsBundle           *BlobsBundle    `json:"blobsBundle,omitempty"`
		ShouldOverrideBuilder *bool           `json:"shouldOverrideBuilder,omitempty"`
	}
	var enc ExecutionPayloadEnvelope
	enc.ExecutionPayload = e.ExecutionPayload
	enc.BlockValue = (*hexutil.Big)(e.BlockValue)
	enc.BlobsBundle = e.BlobsBundle
	enc.ShouldOverrideBuilder = e.ShouldOverrideBuilder
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (e *ExecutionPayloadEnvelope) UnmarshalJSON(input []byte) error {
	type ExecutionPayloadEnvelope struct {
		ExecutionPayload      *ExecutableData `json:"executionPayload"       gencodec:"required"`
		BlockValue            *hexutil.Big    `json:"blockValue"             gencodec:"required"`
		BlobsBundle           *BlobsBundle    `json:"blobsBundle,omitempty"`
		ShouldOverrideBuilder *bool           `json:"shouldOverrideBuilder,omitempty"`
	}
	var dec ExecutionPayloadEnvelope
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.ExecutionPayload == nil {
		return errors.New("missing required field 'executionPayload' for ExecutionPayloadEnvelope")
	}
	e.ExecutionPayload = dec.ExecutionPayload
	if dec.BlockValue == nil {
		return errors.New("missing required field 'blockValue' for ExecutionPayloadEnvelope")
	}
	e.BlockValue = (*big.Int)(dec.BlockValue)
	if dec.BlobsBundle != nil {
		e.BlobsBundle = dec.BlobsBundle
	}
	if dec.ShouldOverrideBuilder != nil {
		e.ShouldOverrideBuilder = dec.ShouldOverrideBuilder
	}
	return nil
}
