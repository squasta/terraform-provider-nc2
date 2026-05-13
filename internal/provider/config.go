// Package provider wires the framework Provider, Resource, DataSource,
// and Action constructors registered by the per-package init blocks.
//
// This file (config.go) implements FR-001a credential resolution:
//
//	precedence: provider block > environment > credentials_file
//
// where the credentials_file is a tiny INI/TOML-shaped file at
// ~/.nc2/credentials by default. When the same credential attribute
// is supplied by more than one source AND the values differ, a
// non-fatal warning diagnostic is emitted naming both sources.
package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/nutanix/terraform-provider-nc2/internal/auth"
)

// EnvAPIKey is the well-known env var supplying api_key.
const EnvAPIKey = "NC2_API_KEY"

// EnvKeyID is the env var supplying key_id.
const EnvKeyID = "NC2_KEY_ID"

// EnvIssuer is the env var supplying issuer.
const EnvIssuer = "NC2_ISSUER"

// EnvProfile is the env var selecting the credentials_file profile.
const EnvProfile = "NC2_PROFILE"

// EnvCredentialsFile is the env var supplying credentials_file.
const EnvCredentialsFile = "NC2_CREDENTIALS_FILE"

// DefaultProfile is the credentials-file section name used when no
// profile is supplied.
const DefaultProfile = "default"

// ProviderBlock captures every provider-block attribute that
// participates in credential resolution. Empty strings mean
// "not set in the block".
type ProviderBlock struct {
	APIKey          string
	KeyID           string
	Issuer          string
	CredentialsFile string
	Profile         string
	BaseURL         string
	CABundle        string
	TaskPollSeconds int64
	TaskMaxSeconds  int64
}

// Env captures the subset of process env vars that participate in
// credential resolution. Tests inject a fake; production calls
// EnvFromOS.
type Env struct {
	APIKey          string
	KeyID           string
	Issuer          string
	Profile         string
	CredentialsFile string
}

// EnvFromOS reads the FR-001a env vars from os.Getenv. Pure-by-input.
func EnvFromOS() Env {
	return Env{
		APIKey:          os.Getenv(EnvAPIKey),
		KeyID:           os.Getenv(EnvKeyID),
		Issuer:          os.Getenv(EnvIssuer),
		Profile:         os.Getenv(EnvProfile),
		CredentialsFile: os.Getenv(EnvCredentialsFile),
	}
}

// File is the in-memory shape of the credentials_file. Parsed from
// the on-disk INI/TOML format by ReadFile (or supplied directly by
// unit tests). Map key = profile name; value = credentials triple.
type File struct {
	Profiles map[string]FileProfile
}

// FileProfile is one [profile] section of the credentials file.
type FileProfile struct {
	APIKey string
	KeyID  string
	Issuer string
}

// ProviderConfig is the fully-resolved set of values the provider
// hands to the runtime layer.
type ProviderConfig struct {
	Credentials      auth.Credentials
	BaseURL          string
	CABundle         string
	TaskPollSeconds  int64
	TaskMaxSeconds   int64
	ProfileUsed      string
	CredentialsFile  string
}

// Resolve performs FR-001a precedence: provider block > env > file.
// On success returns a valid ProviderConfig and the diagnostics
// (which MAY contain non-fatal warnings about source disagreement).
// On missing credentials returns an error diagnostic.
//
// Pure function: all I/O is supplied via the File argument; callers
// pre-load the file via ReadFile.
func Resolve(_ context.Context, block ProviderBlock, env Env, file File) (ProviderConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	profile := firstNonEmpty(block.Profile, env.Profile, DefaultProfile)
	credPath := firstNonEmpty(block.CredentialsFile, env.CredentialsFile)

	fileProf, _ := file.Profiles[profile]

	apiKey := pick("api_key", &diags, src{name: "block", val: block.APIKey}, src{name: "env (NC2_API_KEY)", val: env.APIKey}, src{name: "credentials_file", val: fileProf.APIKey})
	keyID := pick("key_id", &diags, src{name: "block", val: block.KeyID}, src{name: "env (NC2_KEY_ID)", val: env.KeyID}, src{name: "credentials_file", val: fileProf.KeyID})
	issuer := pick("issuer", &diags, src{name: "block", val: block.Issuer}, src{name: "env (NC2_ISSUER)", val: env.Issuer}, src{name: "credentials_file", val: fileProf.Issuer})

	creds := auth.Credentials{APIKey: apiKey, KeyID: keyID, Issuer: issuer}
	if err := creds.Validate(); err != nil {
		diags.AddError("nc2 provider: missing credentials",
			fmt.Sprintf("%v. Supply api_key, key_id, and issuer via the provider block, %s/%s/%s env vars, or the credentials_file (~/.nc2/credentials by default).",
				err, EnvAPIKey, EnvKeyID, EnvIssuer))
		return ProviderConfig{}, diags
	}

	pollSecs := block.TaskPollSeconds
	if pollSecs == 0 {
		pollSecs = 60
	}
	maxSecs := block.TaskMaxSeconds
	if maxSecs == 0 {
		maxSecs = 7200
	}

	return ProviderConfig{
		Credentials:     creds,
		BaseURL:         strings.TrimRight(firstNonEmpty(block.BaseURL, "https://cloud.nutanix.com/api/v2"), "/"),
		CABundle:        block.CABundle,
		TaskPollSeconds: pollSecs,
		TaskMaxSeconds:  maxSecs,
		ProfileUsed:     profile,
		CredentialsFile: credPath,
	}, diags
}

type src struct {
	name string
	val  string
}

// pick implements the FR-001a precedence rule: returns the first
// non-empty source value. If multiple non-empty values disagree, a
// warning diagnostic is emitted naming each source.
func pick(attr string, diags *diag.Diagnostics, sources ...src) string {
	var winner src
	var seen []src
	for _, s := range sources {
		if s.val == "" {
			continue
		}
		seen = append(seen, s)
		if winner.val == "" {
			winner = s
		}
	}
	if len(seen) > 1 {
		var differ []src
		for _, s := range seen {
			if s.val != winner.val {
				differ = append(differ, s)
			}
		}
		if len(differ) > 0 {
			names := []string{winner.name}
			for _, s := range differ {
				names = append(names, s.name)
			}
			diags.AddWarning(
				fmt.Sprintf("nc2 provider: %s supplied by multiple sources with conflicting values", attr),
				fmt.Sprintf("sources: %s. Using value from %s per FR-001a precedence (block > env > credentials_file).",
					strings.Join(names, ", "), winner.name),
			)
		}
	}
	return winner.val
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ReadFile parses an INI-style credentials file. Format:
//
//	[default]
//	api_key = ...
//	key_id  = ...
//	issuer  = ...
//
//	[staging]
//	api_key = ...
//	...
//
// Empty path = "use default ~/.nc2/credentials when present;
// otherwise return an empty File with no error". Unreadable explicit
// path is an error.
func ReadFile(path string) (File, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return File{Profiles: map[string]FileProfile{}}, nil
		}
		path = filepath.Join(home, ".nc2", "credentials")
		_, err = os.Stat(path)
		if err != nil {
			return File{Profiles: map[string]FileProfile{}}, nil
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("provider: read credentials_file %s: %w", path, err)
	}
	return parseINI(data)
}

func parseINI(data []byte) (File, error) {
	out := File{Profiles: map[string]FileProfile{}}
	current := DefaultProfile
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.TrimSpace(line[1 : len(line)-1])
			if current == "" {
				return out, errors.New("provider: empty profile name in credentials_file")
			}
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:eq])
		v := strings.Trim(strings.TrimSpace(line[eq+1:]), "\"'")
		prof := out.Profiles[current]
		switch k {
		case "api_key":
			prof.APIKey = v
		case "key_id":
			prof.KeyID = v
		case "issuer":
			prof.Issuer = v
		}
		out.Profiles[current] = prof
	}
	return out, nil
}
