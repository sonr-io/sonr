// Package app contains tool imports to ensure required dependencies are included
// in the module graph even if they're not directly referenced in the code.
// This prevents "go mod tidy" from removing necessary indirect dependencies.
package app

import (
	_ "cosmossdk.io/orm"
)
