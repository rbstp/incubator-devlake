/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"testing"

	"github.com/apache/incubator-devlake/plugins/argocd/models"
	"github.com/stretchr/testify/assert"
)

func TestParseImageCommitSHA(t *testing.T) {
	cases := []struct {
		name    string
		image   string
		wantSHA string
		wantOK  bool
	}{
		// Rejections.
		{"empty", "", "", false},
		{"whitespace only", "   ", "", false},
		{"no tag", "nginx", "", false},
		{"latest tag", "nginx:latest", "", false},
		{"semver tag", "nginx:1.21", "", false},
		{"three-part semver", "nginx:1.21.0", "", false},
		{"all-numeric short SHA accepted", "myapp:1234567", "1234567", true},
		{"all-numeric tail SHA accepted", "tacet-api:b1433947-1433947", "1433947", true},
		{"bare numeric short SHA accepted", "devolutions/tacet-api:1433947", "1433947", true},
		{"hex below threshold", "myapp:abcdef", "", false},
		{"port without tag", "registry.example.com:5000/myapp", "", false},
		{"digest only", "nginx@sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", "", false},
		{"semver tag with digest", "nginx:1.21@sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", "", false},
		{"13-hex exact accepted", "myapp:abcdef0123456", "abcdef0123456", true},

		// 40-hex matches.
		{"bare 40-hex tag", "myapp:5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", true},
		{"40-hex uppercase", "myapp:5DD95B4EFD7E9B668C361BBDDB8D7F1E56C32AC1", "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", true},
		{"40-hex with prefix", "myapp:sha-5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", true},
		{"40-hex after semver", "myapp:v1.2.3-5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", true},
		{"40-hex with trailing separator", "myapp:5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1-rc1", "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", true},
		{"40-hex stripped of digest", "myapp:5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1@sha256:0000000000000000000000000000000000000000000000000000000000000000", "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", true},
		{"42-hex run rejected", "myapp:5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1ff", "", false},

		// Short hex matches.
		{"exact 7-hex tag", "myapp:abc1234", "abc1234", true},
		{"exact 7-hex uppercase", "myapp:ABC1234", "abc1234", true},
		{"tail short SHA after dash", "myapp:v1.2.3-abc1234", "abc1234", true},
		{"tail short SHA after dot", "myapp:1.2.3.abc1234", "abc1234", true},
		{"tail short SHA after underscore", "myapp:1.2.3_abc1234", "abc1234", true},
		{"tail short SHA after main", "myapp:main-abc1234", "abc1234", true},
		{"head short SHA before dash", "myapp:abc1234-main", "abc1234", true},
		{"head short SHA before dot", "myapp:abc1234.main", "abc1234", true},
		{"16-hex exact tag", "myapp:abcdef0123456789", "abcdef0123456789", true},
		{"12-hex exact tag", "myapp:abcdef012345", "abcdef012345", true},

		// Registry handling.
		{"registry with port and tag", "registry.example.com:5000/myapp:abc1234", "abc1234", true},
		{"registry with port and 40-hex", "registry.example.com:5000/myapp:5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", true},
		{"registry with port no tag", "registry.example.com:5000/myapp", "", false},
		{"deep path with tag", "registry.example.com/team/myapp:abc1234", "abc1234", true},

		// Whitespace.
		{"trimmed whitespace", "  myapp:abc1234  ", "abc1234", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotSHA, gotOK := ParseImageCommitSHA(tc.image)
			assert.Equal(t, tc.wantOK, gotOK, "ok mismatch")
			assert.Equal(t, tc.wantSHA, gotSHA, "sha mismatch")
		})
	}
}

func TestResolveImageRepoUrl_EmptyMappings(t *testing.T) {
	got := resolveImageRepoURL("registry.example.com/myapp:abc1234", nil, "fallback-url")
	assert.Equal(t, "fallback-url", got)
}

func TestResolveImageRepoUrl_SingleMatch(t *testing.T) {
	mappings := []models.ArgocdImageRepoMapping{
		{Pattern: "registry.example.com/payments-*", RepoURL: "https://github.com/example/payments"},
	}
	got := resolveImageRepoURL("registry.example.com/payments-api:abc1234", mappings, "fallback")
	assert.Equal(t, "https://github.com/example/payments", got)
}

func TestResolveImageRepoUrl_NoMatch(t *testing.T) {
	mappings := []models.ArgocdImageRepoMapping{
		{Pattern: "registry.example.com/payments-*", RepoURL: "https://github.com/example/payments"},
	}
	got := resolveImageRepoURL("registry.example.com/billing:abc1234", mappings, "fallback")
	assert.Equal(t, "fallback", got)
}

func TestResolveImageRepoUrl_BadPatternSkipped(t *testing.T) {
	mappings := []models.ArgocdImageRepoMapping{
		{Pattern: "[", RepoURL: "https://github.com/example/bogus"},
		{Pattern: "registry.example.com/billing*", RepoURL: "https://github.com/example/billing"},
	}
	got := resolveImageRepoURL("registry.example.com/billing:abc1234", mappings, "fallback")
	assert.Equal(t, "https://github.com/example/billing", got)
}

func TestResolveImageRepoUrl_FirstMatchWins(t *testing.T) {
	mappings := []models.ArgocdImageRepoMapping{
		{Pattern: "registry.example.com/payments-*", RepoURL: "https://github.com/example/payments-specific"},
		{Pattern: "registry.example.com/*", RepoURL: "https://github.com/example/catchall"},
	}
	got := resolveImageRepoURL("registry.example.com/payments-api:abc1234", mappings, "fallback")
	assert.Equal(t, "https://github.com/example/payments-specific", got)
}

func TestResolveImageRepoUrl_StripsTagAndDigest(t *testing.T) {
	mappings := []models.ArgocdImageRepoMapping{
		{Pattern: "registry.example.com/myapp", RepoURL: "https://github.com/example/myapp"},
	}
	cases := []string{
		"registry.example.com/myapp:abc1234",
		"registry.example.com/myapp@sha256:0000000000000000000000000000000000000000000000000000000000000000",
		"registry.example.com/myapp:abc1234@sha256:0000000000000000000000000000000000000000000000000000000000000000",
		"registry.example.com/myapp",
	}
	for _, image := range cases {
		t.Run(image, func(t *testing.T) {
			got := resolveImageRepoURL(image, mappings, "fallback")
			assert.Equal(t, "https://github.com/example/myapp", got)
		})
	}
}

func TestResolveImageRepoUrl_EmptyPatternSkipped(t *testing.T) {
	mappings := []models.ArgocdImageRepoMapping{
		{Pattern: "", RepoURL: "https://github.com/example/empty"},
		{Pattern: "registry.example.com/billing", RepoURL: "https://github.com/example/billing"},
	}
	got := resolveImageRepoURL("registry.example.com/billing:abc1234", mappings, "fallback")
	assert.Equal(t, "https://github.com/example/billing", got)
}

func TestResolveImageRepoUrl_EmptyRepoUrlSkipped(t *testing.T) {
	mappings := []models.ArgocdImageRepoMapping{
		{Pattern: "registry.example.com/billing", RepoURL: ""},
		{Pattern: "registry.example.com/billing", RepoURL: "https://github.com/example/billing"},
	}
	got := resolveImageRepoURL("registry.example.com/billing:abc1234", mappings, "fallback")
	assert.Equal(t, "https://github.com/example/billing", got)
}
