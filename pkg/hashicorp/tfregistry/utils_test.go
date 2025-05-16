// +build !integration

package tfregistry

import (
	"net/http"
	"testing"
	"strings"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// --- sendRegistryCall ---

var client = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	},
}

var logger = log.New()

func TestSendRegistryCall_Success_v1(t *testing.T) {
	_, err := sendRegistryCall(client, "GET", "providers/hashicorp/aws", logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSendRegistryCall_Success_v2(t *testing.T) {
	_, err := sendRegistryCall(client, "GET", "provider-docs?filter[provider-version]=6221", logger, "v2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSendRegistryCall_404NotFound_v2(t *testing.T) {
	_, err := sendRegistryCall(client, "GET", "test-uri", logger, "v2")
	if err == nil || err.Error() != "error: 404 Not Found" {
		t.Errorf("expected 404 error, got %v", err)
	}
}

func TestSendRegistryCall_404NotFound_v1(t *testing.T) {
	_, err := sendRegistryCall(client, "GET", "test-uri", logger)
	if err == nil || err.Error() != "error: 404 Not Found" {
		t.Errorf("expected 404 error, got %v", err)
	}
}

// --- UnmarshalTFModulePlural ---

func TestUnmarshalTFModulePlural_Valid(t *testing.T) {
	resp := []byte(`{
		"meta": {},
		"modules": [
			{
				"id": "namespace/name/provider/1.0.0",
				"owner": "owner",
				"namespace": "namespace",
				"name": "name",
				"version": "1.0.0",
				"provider": "provider",
				"description": "A test module",
				"source": "source",
				"tag": "",
				"published_at": "2023-01-01T00:00:00Z",
				"downloads": 1,
				"verified": true
			}
		]
	}`)
	out, err := UnmarshalTFModulePlural(resp, "test-query")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(out) == 0 || !strings.Contains(out, "A test module") {
		t.Errorf("expected output to contain module description, got %q", out)
	}
}

func TestUnmarshalTFModulePlural_NoModules(t *testing.T) {
	resp := []byte(`{"meta": {}, "modules": []}`)
	_, err := UnmarshalTFModulePlural(resp, "test-query")
	if err == nil || !strings.Contains(err.Error(), "no modules found") {
		t.Errorf("expected error about no modules found, got %v", err)
	}
}

func TestUnmarshalTFModulePlural_InvalidJSON(t *testing.T) {
	resp := []byte(`not a json`)
	_, err := UnmarshalTFModulePlural(resp, "test-query")
	if err == nil || !strings.Contains(err.Error(), "unmarshalling modules") {
		t.Errorf("expected unmarshalling error, got %v", err)
	}
}

// --- UnmarshalModuleSingular ---

func TestUnmarshalModuleSingular_ValidAllFields(t *testing.T) {
	resp := []byte(`{
		"id": "namespace/name/provider/1.0.0",
		"owner": "owner",
		"namespace": "namespace",
		"name": "name",
		"version": "1.0.0",
		"provider": "provider",
		"provider_logo_url": "",
		"description": "A test module",
		"source": "source",
		"tag": "",
		"published_at": "2023-01-01T00:00:00Z",
		"downloads": 1,
		"verified": true,
		"root": {
			"path": "",
			"name": "root",
			"readme": "",
			"empty": false,
			"inputs": [
				{"name": "input1", "type": "string", "description": "desc", "default": "val", "required": true}
			],
			"outputs": [
				{"name": "output1", "description": "desc"}
			],
			"dependencies": [],
			"provider_dependencies": [
				{"name": "prov1", "namespace": "ns", "source": "src", "version": "1.0.0"}
			],
			"resources": []
		},
		"submodules": [],
		"examples": [
			{"path": "", "name": "example1", "readme": "example readme", "empty": false, "inputs": [], "outputs": [], "dependencies": [], "provider_dependencies": [], "resources": []}
		],
		"providers": ["provider"],
		"versions": ["1.0.0"],
		"deprecation": null
	}`)
	out, err := UnmarshalModuleSingular(resp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "A test module") {
		t.Errorf("expected output to contain module description, got %q", out)
	}
	if !strings.Contains(out, "input1") {
		t.Errorf("expected output to contain input variable, got %q", out)
	}
	if !strings.Contains(out, "example1") {
		t.Errorf("expected output to contain example name, got %q", out)
	}
}

func TestUnmarshalModuleSingular_EmptySections(t *testing.T) {
	resp := []byte(`{
		"id": "namespace/name/provider/1.0.0",
		"owner": "owner",
		"namespace": "namespace",
		"name": "name",
		"version": "1.0.0",
		"provider": "provider",
		"provider_logo_url": "",
		"description": "A test module",
		"source": "source",
		"tag": "",
		"published_at": "2023-01-01T00:00:00Z",
		"downloads": 1,
		"verified": true,
		"root": {
			"path": "",
			"name": "root",
			"readme": "",
			"empty": false,
			"inputs": [],
			"outputs": [],
			"dependencies": [],
			"provider_dependencies": [],
			"resources": []
		},
		"submodules": [],
		"examples": [],
		"providers": ["provider"],
		"versions": ["1.0.0"],
		"deprecation": null
	}`)
	out, err := UnmarshalModuleSingular(resp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "A test module") {
		t.Errorf("expected output to contain description, got %q", out)
	}
}

func TestUnmarshalModuleSingular_InvalidJSON(t *testing.T) {
	resp := []byte(`not a json`)
	_, err := UnmarshalModuleSingular(resp)
	if err == nil || !strings.Contains(err.Error(), "unmarshalling module details") {
		t.Errorf("expected unmarshalling error, got %v", err)
	}
}

// --- Others---

func TestExtractProviderNameAndVersion(t *testing.T) {
	uri := "registry://providers/hashicorp/namespace/aws/version/3.0.0"
	ns, name, version := ExtractProviderNameAndVersion(uri)
	if ns != "hashicorp" || name != "aws" || version != "3.0.0" {
		t.Errorf("expected (hashicorp, aws, 3.0.0), got (%s, %s, %s)", ns, name, version)
	}
}

func TestConstructProviderVersionURI(t *testing.T) {
	uri := ConstructProviderVersionURI("hashicorp", "aws", "3.0.0")
	expected := "registry://providers/hashicorp/providers/aws/versions/3.0.0"
	if uri != expected {
		t.Errorf("expected %q, got %q", expected, uri)
	}
}

func TestContainsSlug(t *testing.T) {
	ok, err := containsSlug("aws_s3_bucket", "s3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("expected true, got false")
	}
	ok, err = containsSlug("aws_s3_bucket", "ec2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Errorf("expected false, got true")
	}
}

func TestIsValidProviderVersionFormat(t *testing.T) {
	valid := []string{"1.0.0", "v1.2.3", "1.0.0-beta"}
	invalid := []string{"1.0", "v1", "foo", ""}
	for _, v := range valid {
		if !isValidProviderVersionFormat(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	for _, v := range invalid {
		if isValidProviderVersionFormat(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestIsValidProviderDataType(t *testing.T) {
	valid := []string{"resources", "data-sources", "functions", "guides", "overview"}
	invalid := []string{"foo", "bar", ""}
	for _, v := range valid {
		if !isValidProviderDataType(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	for _, v := range invalid {
		if isValidProviderDataType(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestLogAndReturnError_NilLogger(t *testing.T) {
	err := logAndReturnError(nil, "context", fmt.Errorf("fail"))
	if err == nil || !strings.Contains(err.Error(), "context") {
		t.Errorf("expected error to contain context, got %v", err)
	}
}

func TestIsV2ProviderDataType(t *testing.T) {
	valid := []string{"guides", "functions", "overview"}
	invalid := []string{"resources", "data-sources", "foo"}
	for _, v := range valid {
		if !isV2ProviderDataType(v) {
			t.Errorf("expected %q to be valid v2 data type", v)
		}
	}
	for _, v := range invalid {
		if isV2ProviderDataType(v) {
			t.Errorf("expected %q to be invalid v2 data type", v)
		}
	}
}
