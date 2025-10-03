// Package webauthn provides performance optimizations for WebAuthn operations.
package webauthn

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"hash"
	"math/big"
	"sync"
	"time"

	"github.com/sonr-io/sonr/types/webauthn/webauthncose"
)

// Performance optimization structures and caches

// CachedEC2PublicKey stores parsed ECDSA public key for reuse.
type CachedEC2PublicKey struct {
	PublicKey *ecdsa.PublicKey
	Algorithm webauthncose.COSEAlgorithmIdentifier
}

// CredentialCache provides thread-safe caching for parsed credentials.
type CredentialCache struct {
	pubkeys    sync.Map // map[string]*CachedEC2PublicKey
	challenges sync.Map // map[string]*ChallengeValidation
	hashes     sync.Map // map[string][32]byte
}

// ChallengeValidation stores cached challenge validation results.
type ChallengeValidation struct {
	IsValid   bool
	ExpiresAt time.Time
}

var (
	// Global credential cache instance
	globalCache = &CredentialCache{}

	// Hash pool for reusing hash instances
	hashPool = sync.Pool{
		New: func() any {
			return sha256.New()
		},
	}
)

// GetCache returns the global credential cache.
func GetCache() *CredentialCache {
	return globalCache
}

// GetCachedPublicKey retrieves or creates a cached ECDSA public key.
func (c *CredentialCache) GetCachedPublicKey(
	xCoord, yCoord []byte,
	algorithm int64,
) (*CachedEC2PublicKey, error) {
	cacheKey := fmt.Sprintf("%x-%x", xCoord, yCoord)

	// Try to load from cache
	if cached, ok := c.pubkeys.Load(cacheKey); ok {
		return cached.(*CachedEC2PublicKey), nil
	}

	// Create new cached key
	curve := webauthncose.EC2AlgCurve(algorithm)
	if curve == nil {
		return nil, webauthncose.ErrUnsupportedAlgorithm
	}

	cachedKey := &CachedEC2PublicKey{
		PublicKey: &ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(xCoord),
			Y:     new(big.Int).SetBytes(yCoord),
		},
		Algorithm: webauthncose.COSEAlgorithmIdentifier(algorithm),
	}

	// Store in cache
	actual, _ := c.pubkeys.LoadOrStore(cacheKey, cachedKey)
	return actual.(*CachedEC2PublicKey), nil
}

// VerifyWithCache performs ECDSA verification using cached public key.
func (c *CredentialCache) VerifyWithCache(
	key *webauthncose.EC2PublicKeyData,
	data []byte,
	sig []byte,
) (bool, error) {
	// Get cached public key
	cached, err := c.GetCachedPublicKey(key.XCoord, key.YCoord, key.Algorithm)
	if err != nil {
		return false, err
	}

	// Get hash from pool
	h := hashPool.Get().(hash.Hash)
	defer func() {
		h.Reset()
		hashPool.Put(h)
	}()

	h.Write(data)
	hashed := h.Sum(nil)

	// Parse signature
	type ECDSASignature struct {
		R, S *big.Int
	}
	var e ECDSASignature
	if _, err := asn1.Unmarshal(sig, &e); err != nil {
		return false, webauthncose.ErrSigNotProvidedOrInvalid
	}

	return ecdsa.Verify(cached.PublicKey, hashed, e.R, e.S), nil
}

// GetClientDataHash returns cached SHA256 hash of client data.
func (c *CredentialCache) GetClientDataHash(clientDataJSON []byte) [32]byte {
	// Create cache key from first 8 bytes of hash
	quickHash := sha256.Sum256(clientDataJSON)
	cacheKey := fmt.Sprintf("%x", quickHash[:8])

	// Check cache
	if cached, ok := c.hashes.Load(cacheKey); ok {
		return cached.([32]byte)
	}

	// Store and return
	c.hashes.Store(cacheKey, quickHash)
	return quickHash
}

// ValidateChallengeWithCache validates challenge with caching.
func (c *CredentialCache) ValidateChallengeWithCache(
	challenge string,
	expected []byte,
) (bool, error) {
	cacheKey := fmt.Sprintf("%s-%x", challenge, expected)

	// Check cache
	if cached, ok := c.challenges.Load(cacheKey); ok {
		cv := cached.(*ChallengeValidation)
		if time.Now().Before(cv.ExpiresAt) {
			return cv.IsValid, nil
		}
		// Expired, remove from cache
		c.challenges.Delete(cacheKey)
	}

	// Perform validation
	isValid := challenge == fmt.Sprintf("%x", expected)

	// Cache result for 5 minutes
	c.challenges.Store(cacheKey, &ChallengeValidation{
		IsValid:   isValid,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})

	return isValid, nil
}

// CleanExpiredChallenges removes expired challenge validations from cache.
func (c *CredentialCache) CleanExpiredChallenges() {
	now := time.Now()
	c.challenges.Range(func(key, value any) bool {
		cv := value.(*ChallengeValidation)
		if now.After(cv.ExpiresAt) {
			c.challenges.Delete(key)
		}
		return true
	})
}

// StartCacheCleaner starts a background goroutine to clean expired entries.
func StartCacheCleaner(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			globalCache.CleanExpiredChallenges()
		}
	}()
}

// ClearCache removes all cached entries.
func (c *CredentialCache) ClearCache() {
	c.pubkeys.Range(func(key, value any) bool {
		c.pubkeys.Delete(key)
		return true
	})

	c.challenges.Range(func(key, value any) bool {
		c.challenges.Delete(key)
		return true
	})

	c.hashes.Range(func(key, value any) bool {
		c.hashes.Delete(key)
		return true
	})
}

// CacheStats returns cache statistics.
type CacheStats struct {
	PublicKeys int
	Challenges int
	Hashes     int
}

// GetStats returns current cache statistics.
func (c *CredentialCache) GetStats() CacheStats {
	stats := CacheStats{}

	c.pubkeys.Range(func(key, value any) bool {
		stats.PublicKeys++
		return true
	})

	c.challenges.Range(func(key, value any) bool {
		stats.Challenges++
		return true
	})

	c.hashes.Range(func(key, value any) bool {
		stats.Hashes++
		return true
	})

	return stats
}
