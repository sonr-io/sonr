package types

import (
	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
)

// ToAPIEncryptionMetadata converts types.EncryptionMetadata to apiv1.EncryptionMetadata
func (em *EncryptionMetadata) ToAPIEncryptionMetadata() *apiv1.EncryptionMetadata {
	if em == nil {
		return nil
	}

	return &apiv1.EncryptionMetadata{
		Algorithm:        em.Algorithm,
		ConsensusInput:   em.ConsensusInput,
		Nonce:            em.Nonce,
		AuthTag:          em.AuthTag,
		EncryptionHeight: em.EncryptionHeight,
		ValidatorSet:     em.ValidatorSet,
		KeyVersion:       em.KeyVersion,
		SingleNodeMode:   em.SingleNodeMode,
	}
}

// FromAPIEncryptionMetadata converts apiv1.EncryptionMetadata to types.EncryptionMetadata
func FromAPIEncryptionMetadata(apiMeta *apiv1.EncryptionMetadata) *EncryptionMetadata {
	if apiMeta == nil {
		return nil
	}

	return &EncryptionMetadata{
		Algorithm:        apiMeta.Algorithm,
		ConsensusInput:   apiMeta.ConsensusInput,
		Nonce:            apiMeta.Nonce,
		AuthTag:          apiMeta.AuthTag,
		EncryptionHeight: apiMeta.EncryptionHeight,
		ValidatorSet:     apiMeta.ValidatorSet,
		KeyVersion:       apiMeta.KeyVersion,
		SingleNodeMode:   apiMeta.SingleNodeMode,
	}
}

// ToAPIVRFContribution converts types.VRFContribution to apiv1.VRFContribution
func (vrf *VRFContribution) ToAPIVRFContribution() *apiv1.VRFContribution {
	if vrf == nil {
		return nil
	}

	return &apiv1.VRFContribution{
		ValidatorAddress: vrf.ValidatorAddress,
		Randomness:       vrf.Randomness,
		Proof:            vrf.Proof,
		BlockHeight:      vrf.BlockHeight,
		Timestamp:        vrf.Timestamp,
	}
}

// FromAPIVRFContribution converts apiv1.VRFContribution to types.VRFContribution
func FromAPIVRFContribution(apiVrf *apiv1.VRFContribution) *VRFContribution {
	if apiVrf == nil {
		return nil
	}

	return &VRFContribution{
		ValidatorAddress: apiVrf.ValidatorAddress,
		Randomness:       apiVrf.Randomness,
		Proof:            apiVrf.Proof,
		BlockHeight:      apiVrf.BlockHeight,
		Timestamp:        apiVrf.Timestamp,
	}
}

// ToAPIEncryptionKeyState converts types.EncryptionKeyState to apiv1.EncryptionKeyState
func (eks *EncryptionKeyState) ToAPIEncryptionKeyState() *apiv1.EncryptionKeyState {
	if eks == nil {
		return nil
	}

	// Convert contribution slice to API format
	apiContributions := make([]*apiv1.VRFContribution, len(eks.Contributions))
	for i, contrib := range eks.Contributions {
		if contrib != nil {
			apiContributions[i] = contrib.ToAPIVRFContribution()
		}
	}

	return &apiv1.EncryptionKeyState{
		CurrentKey:     eks.CurrentKey,
		KeyVersion:     eks.KeyVersion,
		ValidatorSet:   eks.ValidatorSet,
		Contributions:  apiContributions,
		LastRotation:   eks.LastRotation,
		NextRotation:   eks.NextRotation,
		SingleNodeMode: eks.SingleNodeMode,
	}
}

// FromAPIEncryptionKeyState converts apiv1.EncryptionKeyState to types.EncryptionKeyState
func FromAPIEncryptionKeyState(apiEks *apiv1.EncryptionKeyState) *EncryptionKeyState {
	if apiEks == nil {
		return nil
	}

	// Convert API contribution slice to types format
	contributions := make([]*VRFContribution, len(apiEks.Contributions))
	for i, apiContrib := range apiEks.Contributions {
		if apiContrib != nil {
			contributions[i] = FromAPIVRFContribution(apiContrib)
		}
	}

	return &EncryptionKeyState{
		CurrentKey:     apiEks.CurrentKey,
		KeyVersion:     apiEks.KeyVersion,
		ValidatorSet:   apiEks.ValidatorSet,
		Contributions:  contributions,
		LastRotation:   apiEks.LastRotation,
		NextRotation:   apiEks.NextRotation,
		SingleNodeMode: apiEks.SingleNodeMode,
	}
}

// ToAPIVRFConsensusRound converts types.VRFConsensusRound to apiv1.VRFConsensusRound
func (vcr *VRFConsensusRound) ToAPIVRFConsensusRound() *apiv1.VRFConsensusRound {
	if vcr == nil {
		return nil
	}

	return &apiv1.VRFConsensusRound{
		RoundNumber:           vcr.RoundNumber,
		KeyVersion:            vcr.KeyVersion,
		RequiredContributions: vcr.RequiredContributions,
		ReceivedContributions: vcr.ReceivedContributions,
		Status:                vcr.Status,
		ExpiryHeight:          vcr.ExpiryHeight,
		InitiatedHeight:       vcr.InitiatedHeight,
		ConsensusInput:        vcr.ConsensusInput,
		Completed:             vcr.Completed,
	}
}

// FromAPIVRFConsensusRound converts apiv1.VRFConsensusRound to types.VRFConsensusRound
func FromAPIVRFConsensusRound(apiVcr *apiv1.VRFConsensusRound) *VRFConsensusRound {
	if apiVcr == nil {
		return nil
	}

	return &VRFConsensusRound{
		RoundNumber:           apiVcr.RoundNumber,
		KeyVersion:            apiVcr.KeyVersion,
		RequiredContributions: apiVcr.RequiredContributions,
		ReceivedContributions: apiVcr.ReceivedContributions,
		Status:                apiVcr.Status,
		ExpiryHeight:          apiVcr.ExpiryHeight,
		InitiatedHeight:       apiVcr.InitiatedHeight,
		ConsensusInput:        apiVcr.ConsensusInput,
		Completed:             apiVcr.Completed,
	}
}

// ToAPIEncryptionStats converts types.EncryptionStats to apiv1.EncryptionStats
func (es *EncryptionStats) ToAPIEncryptionStats() *apiv1.EncryptionStats {
	if es == nil {
		return nil
	}

	return &apiv1.EncryptionStats{
		TotalEncryptedRecords: es.TotalEncryptedRecords,
		TotalDecryptionErrors: es.TotalDecryptionErrors,
		LastEncryptionHeight:  es.LastEncryptionHeight,
	}
}

// FromAPIEncryptionStats converts apiv1.EncryptionStats to types.EncryptionStats
func FromAPIEncryptionStats(apiEs *apiv1.EncryptionStats) *EncryptionStats {
	if apiEs == nil {
		return nil
	}

	return &EncryptionStats{
		TotalEncryptedRecords: apiEs.TotalEncryptedRecords,
		TotalDecryptionErrors: apiEs.TotalDecryptionErrors,
		LastEncryptionHeight:  apiEs.LastEncryptionHeight,
	}
}
