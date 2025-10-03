package types

import (
	apiv1 "github.com/sonr-io/sonr/api/did/v1"
)

// ToORMDIDDocument converts a DIDDocument from the types package to the ORM API type
func (d *DIDDocument) ToORM() *apiv1.DIDDocument {
	if d == nil {
		return nil
	}

	ormDoc := &apiv1.DIDDocument{
		Id:                d.Id,
		PrimaryController: d.PrimaryController,
		AlsoKnownAs:       d.AlsoKnownAs,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
		Deactivated:       d.Deactivated,
		Version:           d.Version,
	}

	// Convert verification methods
	ormDoc.VerificationMethod = make([]*apiv1.VerificationMethod, len(d.VerificationMethod))
	for i, vm := range d.VerificationMethod {
		ormDoc.VerificationMethod[i] = vm.ToORM()
	}

	// Convert verification method references
	ormDoc.Authentication = convertVerificationMethodReferencesToORM(d.Authentication)
	ormDoc.AssertionMethod = convertVerificationMethodReferencesToORM(d.AssertionMethod)
	ormDoc.KeyAgreement = convertVerificationMethodReferencesToORM(d.KeyAgreement)
	ormDoc.CapabilityInvocation = convertVerificationMethodReferencesToORM(d.CapabilityInvocation)
	ormDoc.CapabilityDelegation = convertVerificationMethodReferencesToORM(d.CapabilityDelegation)

	// Convert services
	ormDoc.Service = make([]*apiv1.Service, len(d.Service))
	for i, svc := range d.Service {
		ormDoc.Service[i] = svc.ToORM()
	}

	return ormDoc
}

// ToORMVerificationMethod converts a VerificationMethod from the types package to the ORM API type
func (vm *VerificationMethod) ToORM() *apiv1.VerificationMethod {
	if vm == nil {
		return nil
	}

	ormVM := &apiv1.VerificationMethod{
		Id:                     vm.Id,
		VerificationMethodKind: vm.VerificationMethodKind,
		Controller:             vm.Controller,
		PublicKeyJwk:           vm.PublicKeyJwk,
		PublicKeyMultibase:     vm.PublicKeyMultibase,
		PublicKeyBase58:        vm.PublicKeyBase58,
		PublicKeyBase64:        vm.PublicKeyBase64,
		PublicKeyPem:           vm.PublicKeyPem,
		PublicKeyHex:           vm.PublicKeyHex,
	}

	// Convert WebAuthn credential if present
	if vm.WebauthnCredential != nil {
		ormVM.WebauthnCredential = &apiv1.WebAuthnCredential{
			CredentialId:       vm.WebauthnCredential.CredentialId,
			PublicKey:          vm.WebauthnCredential.PublicKey,
			Algorithm:          vm.WebauthnCredential.Algorithm,
			AttestationType:    vm.WebauthnCredential.AttestationType,
			Origin:             vm.WebauthnCredential.Origin,
			CreatedAt:          vm.WebauthnCredential.CreatedAt,
			RpId:               vm.WebauthnCredential.RpId,
			RpName:             vm.WebauthnCredential.RpName,
			Transports:         vm.WebauthnCredential.Transports,
			UserVerified:       vm.WebauthnCredential.UserVerified,
			SignatureAlgorithm: vm.WebauthnCredential.SignatureAlgorithm,
			RawId:              vm.WebauthnCredential.RawId,
			ClientDataJson:     vm.WebauthnCredential.ClientDataJson,
			AttestationObject:  vm.WebauthnCredential.AttestationObject,
		}
	}

	return ormVM
}

// ToORMService converts a Service from the types package to the ORM API type
func (s *Service) ToORM() *apiv1.Service {
	if s == nil {
		return nil
	}

	ormService := &apiv1.Service{
		Id:              s.Id,
		ServiceKind:     s.ServiceKind,
		SingleEndpoint:  s.SingleEndpoint,
		ComplexEndpoint: s.ComplexEndpoint,
		Properties:      s.Properties,
	}

	// Convert multiple endpoints if present
	if s.MultipleEndpoints != nil {
		ormService.MultipleEndpoints = &apiv1.ServiceEndpoints{
			Endpoints: s.MultipleEndpoints.Endpoints,
		}
	}

	return ormService
}

// ToORMVerifiableCredential converts a VerifiableCredential from the types package to the ORM API type
func (vc *VerifiableCredential) ToORM() *apiv1.VerifiableCredential {
	if vc == nil {
		return nil
	}

	ormVC := &apiv1.VerifiableCredential{
		Id:                vc.Id,
		Context:           vc.Context,
		CredentialKinds:   vc.CredentialKinds,
		Issuer:            vc.Issuer,
		IssuanceDate:      vc.IssuanceDate,
		ExpirationDate:    vc.ExpirationDate,
		CredentialSubject: vc.CredentialSubject,
		Subject:           vc.Subject,
		IssuedAt:          vc.IssuedAt,
		ExpiresAt:         vc.ExpiresAt,
		Revoked:           vc.Revoked,
	}

	// Convert proofs
	ormVC.Proof = make([]*apiv1.CredentialProof, len(vc.Proof))
	for i, proof := range vc.Proof {
		ormVC.Proof[i] = &apiv1.CredentialProof{
			ProofKind:          proof.ProofKind,
			Created:            proof.Created,
			VerificationMethod: proof.VerificationMethod,
			ProofPurpose:       proof.ProofPurpose,
			Signature:          proof.Signature,
			Properties:         proof.Properties,
		}
	}

	// Convert credential status if present
	if vc.CredentialStatus != nil {
		ormVC.CredentialStatus = &apiv1.CredentialStatus{
			Id:         vc.CredentialStatus.Id,
			StatusKind: vc.CredentialStatus.StatusKind,
			Properties: vc.CredentialStatus.Properties,
		}
	}

	return ormVC
}

// ToORMDIDDocumentMetadata converts DIDDocumentMetadata from the types package to the ORM API type
func (m *DIDDocumentMetadata) ToORM() *apiv1.DIDDocumentMetadata {
	if m == nil {
		return nil
	}

	return &apiv1.DIDDocumentMetadata{
		Did:           m.Did,
		Created:       m.Created,
		Updated:       m.Updated,
		Deactivated:   m.Deactivated,
		VersionId:     m.VersionId,
		NextUpdate:    m.NextUpdate,
		NextVersionId: m.NextVersionId,
		EquivalentId:  m.EquivalentId,
		CanonicalId:   m.CanonicalId,
	}
}

// FromORMDIDDocument converts a DIDDocument from the ORM API type to the types package
func DIDDocumentFromORM(ormDoc *apiv1.DIDDocument) *DIDDocument {
	if ormDoc == nil {
		return nil
	}

	doc := &DIDDocument{
		Id:                ormDoc.Id,
		PrimaryController: ormDoc.PrimaryController,
		AlsoKnownAs:       ormDoc.AlsoKnownAs,
		CreatedAt:         ormDoc.CreatedAt,
		UpdatedAt:         ormDoc.UpdatedAt,
		Deactivated:       ormDoc.Deactivated,
		Version:           ormDoc.Version,
	}

	// Convert verification methods
	doc.VerificationMethod = make([]*VerificationMethod, len(ormDoc.VerificationMethod))
	for i, vm := range ormDoc.VerificationMethod {
		doc.VerificationMethod[i] = VerificationMethodFromORM(vm)
	}

	// Convert verification method references
	doc.Authentication = convertVerificationMethodReferencesFromORM(ormDoc.Authentication)
	doc.AssertionMethod = convertVerificationMethodReferencesFromORM(ormDoc.AssertionMethod)
	doc.KeyAgreement = convertVerificationMethodReferencesFromORM(ormDoc.KeyAgreement)
	doc.CapabilityInvocation = convertVerificationMethodReferencesFromORM(
		ormDoc.CapabilityInvocation,
	)
	doc.CapabilityDelegation = convertVerificationMethodReferencesFromORM(
		ormDoc.CapabilityDelegation,
	)

	// Convert services
	doc.Service = make([]*Service, len(ormDoc.Service))
	for i, svc := range ormDoc.Service {
		doc.Service[i] = ServiceFromORM(svc)
	}

	return doc
}

// VerificationMethodFromORM converts a VerificationMethod from the ORM API type to the types package
func VerificationMethodFromORM(ormVM *apiv1.VerificationMethod) *VerificationMethod {
	if ormVM == nil {
		return nil
	}

	vm := &VerificationMethod{
		Id:                     ormVM.Id,
		VerificationMethodKind: ormVM.VerificationMethodKind,
		Controller:             ormVM.Controller,
		PublicKeyJwk:           ormVM.PublicKeyJwk,
		PublicKeyMultibase:     ormVM.PublicKeyMultibase,
		PublicKeyBase58:        ormVM.PublicKeyBase58,
		PublicKeyBase64:        ormVM.PublicKeyBase64,
		PublicKeyPem:           ormVM.PublicKeyPem,
		PublicKeyHex:           ormVM.PublicKeyHex,
	}

	// Convert WebAuthn credential if present
	if ormVM.WebauthnCredential != nil {
		vm.WebauthnCredential = &WebAuthnCredential{
			CredentialId:       ormVM.WebauthnCredential.CredentialId,
			PublicKey:          ormVM.WebauthnCredential.PublicKey,
			Algorithm:          ormVM.WebauthnCredential.Algorithm,
			AttestationType:    ormVM.WebauthnCredential.AttestationType,
			Origin:             ormVM.WebauthnCredential.Origin,
			CreatedAt:          ormVM.WebauthnCredential.CreatedAt,
			RpId:               ormVM.WebauthnCredential.RpId,
			RpName:             ormVM.WebauthnCredential.RpName,
			Transports:         ormVM.WebauthnCredential.Transports,
			UserVerified:       ormVM.WebauthnCredential.UserVerified,
			SignatureAlgorithm: ormVM.WebauthnCredential.SignatureAlgorithm,
			RawId:              ormVM.WebauthnCredential.RawId,
			ClientDataJson:     ormVM.WebauthnCredential.ClientDataJson,
			AttestationObject:  ormVM.WebauthnCredential.AttestationObject,
		}
	}

	return vm
}

// ServiceFromORM converts a Service from the ORM API type to the types package
func ServiceFromORM(ormService *apiv1.Service) *Service {
	if ormService == nil {
		return nil
	}

	svc := &Service{
		Id:              ormService.Id,
		ServiceKind:     ormService.ServiceKind,
		SingleEndpoint:  ormService.SingleEndpoint,
		ComplexEndpoint: ormService.ComplexEndpoint,
		Properties:      ormService.Properties,
	}

	// Convert multiple endpoints if present
	if ormService.MultipleEndpoints != nil {
		svc.MultipleEndpoints = &ServiceEndpoints{
			Endpoints: ormService.MultipleEndpoints.Endpoints,
		}
	}

	return svc
}

// VerifiableCredentialFromORM converts a VerifiableCredential from the ORM API type to the types package
func VerifiableCredentialFromORM(ormVC *apiv1.VerifiableCredential) *VerifiableCredential {
	if ormVC == nil {
		return nil
	}

	vc := &VerifiableCredential{
		Id:                ormVC.Id,
		Context:           ormVC.Context,
		CredentialKinds:   ormVC.CredentialKinds,
		Issuer:            ormVC.Issuer,
		IssuanceDate:      ormVC.IssuanceDate,
		ExpirationDate:    ormVC.ExpirationDate,
		CredentialSubject: ormVC.CredentialSubject,
		Subject:           ormVC.Subject,
		IssuedAt:          ormVC.IssuedAt,
		ExpiresAt:         ormVC.ExpiresAt,
		Revoked:           ormVC.Revoked,
	}

	// Convert proofs
	vc.Proof = make([]*CredentialProof, len(ormVC.Proof))
	for i, proof := range ormVC.Proof {
		vc.Proof[i] = &CredentialProof{
			ProofKind:          proof.ProofKind,
			Created:            proof.Created,
			VerificationMethod: proof.VerificationMethod,
			ProofPurpose:       proof.ProofPurpose,
			Signature:          proof.Signature,
			Properties:         proof.Properties,
		}
	}

	// Convert credential status if present
	if ormVC.CredentialStatus != nil {
		vc.CredentialStatus = &CredentialStatus{
			Id:         ormVC.CredentialStatus.Id,
			StatusKind: ormVC.CredentialStatus.StatusKind,
			Properties: ormVC.CredentialStatus.Properties,
		}
	}

	return vc
}

// DIDDocumentMetadataFromORM converts DIDDocumentMetadata from the ORM API type to the types package
func DIDDocumentMetadataFromORM(ormMeta *apiv1.DIDDocumentMetadata) *DIDDocumentMetadata {
	if ormMeta == nil {
		return nil
	}

	return &DIDDocumentMetadata{
		Did:           ormMeta.Did,
		Created:       ormMeta.Created,
		Updated:       ormMeta.Updated,
		Deactivated:   ormMeta.Deactivated,
		VersionId:     ormMeta.VersionId,
		NextUpdate:    ormMeta.NextUpdate,
		NextVersionId: ormMeta.NextVersionId,
		EquivalentId:  ormMeta.EquivalentId,
		CanonicalId:   ormMeta.CanonicalId,
	}
}

// Helper functions

func convertVerificationMethodReferencesToORM(
	refs []*VerificationMethodReference,
) []*apiv1.VerificationMethodReference {
	if refs == nil {
		return nil
	}

	ormRefs := make([]*apiv1.VerificationMethodReference, len(refs))
	for i, ref := range refs {
		if ref == nil {
			continue
		}

		ormRef := &apiv1.VerificationMethodReference{}

		if ref.VerificationMethodId != "" {
			ormRef.VerificationMethodId = ref.VerificationMethodId
		} else if ref.EmbeddedVerificationMethod != nil {
			ormRef.EmbeddedVerificationMethod = ref.EmbeddedVerificationMethod.ToORM()
		}

		ormRefs[i] = ormRef
	}

	return ormRefs
}

func convertVerificationMethodReferencesFromORM(
	ormRefs []*apiv1.VerificationMethodReference,
) []*VerificationMethodReference {
	if ormRefs == nil {
		return nil
	}

	refs := make([]*VerificationMethodReference, len(ormRefs))
	for i, ormRef := range ormRefs {
		if ormRef == nil {
			continue
		}

		ref := &VerificationMethodReference{}

		if ormRef.VerificationMethodId != "" {
			ref.VerificationMethodId = ormRef.VerificationMethodId
		} else if ormRef.EmbeddedVerificationMethod != nil {
			ref.EmbeddedVerificationMethod = VerificationMethodFromORM(ormRef.EmbeddedVerificationMethod)
		}

		refs[i] = ref
	}

	return refs
}
