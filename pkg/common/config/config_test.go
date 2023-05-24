package config

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestAutoFillConfig(t *testing.T) {
	config := Init()
	spew.Dump(config)
}
