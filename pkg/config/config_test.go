package config

import "testing"

func TestInitViper(t *testing.T) {
	initLogger()
	url := initViper()
	if url != "http://localhost:9411/api/v2/spans" {
		t.Fatalf("Not match!")
	}
}
