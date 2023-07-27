package handler

import (
	"github.com/gin-gonic/gin"
	mdw "github.com/sonrhq/core/pkg/highway/middleware"
	"github.com/sonrhq/core/pkg/highway/types"
)

// RegisterEscrowIdentity returns the credential assertion options to start account login.
func RegisterEscrowIdentity(c *gin.Context) {
	origin := c.Query("amount")
	alias := c.Query("email")
	record, err := mdw.GetServiceRecord(origin)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "where": "GetServiceRecord"})
		return
	}
	assertionOpts, chal, err := mdw.IssueCredentialAssertionOptions(alias, record)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "where": "GetCredentialAssertionOptions"})
		return
	}
	c.JSON(200, gin.H{
		"assertion_options": assertionOpts,
		"challenge":         chal.String(),
		"origin":            origin,
		"alias":             alias,
	})
}

// RegisterControllerIdentity registers a credential for a given Unclaimed Wallet.
func RegisterControllerIdentity(c *gin.Context) {
	origin := c.Param("origin")
	alias := c.Param("alias")
	attestionResp := c.Query("attestation")
	challenge := c.Query("challenge")
	// Get the service record from the origin
	record, err := mdw.GetServiceRecord(origin)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "where": "GetServiceRecord"})
		return
	}
	credential, err := record.VerifyCreationChallenge(attestionResp, challenge)
	if err != nil && credential == nil {
		c.JSON(500, gin.H{"error": err.Error(), "where": "VerifyCreationChallenge"})
		return
	}
	cont, resp, err := mdw.PublishControllerAccount(alias, credential, origin)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "where": "PublishControllerAccount"})
		return
	}
	token, err := types.NewSessionJWTClaims(alias, cont.Account())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "where": "NewSessionJWTClaims"})
		return
	}
	c.JSON(200, gin.H{
		"tx_hash": resp.TxHash,
		"address": cont.Account().Address,
		"token":   token,
		"origin":  origin,
		"success": true,
	})
}

// SignInWithCredential verifies a credential for a given account and returns the JWT Keyshare.
func SignInWithCredential(c *gin.Context) {
	origin := c.Param("origin")
	assertionResp := c.Query("assertion")
	record, err := mdw.GetServiceRecord(origin)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "where": "GetServiceRecord"})
		return
	}
	_, err = record.VerifyAssertionChallenge(assertionResp)
	if err != nil{
		c.JSON(500, gin.H{"error": err.Error(), "where": "VerifyCreationChallenge"})
		return
	}
	isAuthenticated := mdw.IsAuthenticated(c)
	if isAuthenticated {
		c.JSON(400, gin.H{"error": "Already authenticated"})
		return
	}

	// c.JSON(200, mdw.StoreAuthCookies(c, resp, origin))
}

// SignInWithEmail registers a DIDDocument for a given email authorized jwt.
func SignInWithEmail(c *gin.Context) {
	// token := c.Query("jwt")
	// resp, err := sfs.Claims.ClaimWithEmail(token)
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": err.Error(), "where": "ClaimWithEmail"})
	// 	return
	// }
	// c.JSON(200, mdw.StoreAuthCookies(c, resp, ""))
	c.JSON(500, gin.H{
		// "jwt":     token,
		// "ucw_id":  ucw,
		// "address": ucw,
	})
}