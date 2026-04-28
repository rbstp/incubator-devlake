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
	"path"
	"regexp"
	"strings"

	"github.com/apache/incubator-devlake/plugins/argocd/models"
)

var (
	imageSHA40Pattern    = regexp.MustCompile(`(?i)([0-9a-f]{40})`)
	imageSHATailPattern  = regexp.MustCompile(`(?i)[._-]([0-9a-f]{7,12})$`)
	imageSHAHeadPattern  = regexp.MustCompile(`(?i)^([0-9a-f]{7,12})[._-]`)
	imageSHAExactPattern = regexp.MustCompile(`(?i)^([0-9a-f]{7,40})$`)
)

// ParseImageCommitSHA recovers a git commit SHA from an image tag, returning
// ("", false) when none is present. The image content digest (@sha256:...) is
// stripped first so it is never returned as a source SHA.
func ParseImageCommitSHA(image string) (string, bool) {
	image = stripImageDigest(strings.TrimSpace(image))
	if image == "" {
		return "", false
	}

	tag := extractImageTag(image)
	if tag == "" {
		return "", false
	}

	if locs := imageSHA40Pattern.FindStringIndex(tag); locs != nil {
		if hasHexBoundaries(tag, locs[0], locs[1]) {
			return strings.ToLower(tag[locs[0]:locs[1]]), true
		}
	}

	for _, re := range []*regexp.Regexp{imageSHATailPattern, imageSHAHeadPattern, imageSHAExactPattern} {
		if m := re.FindStringSubmatch(tag); m != nil {
			return strings.ToLower(m[1]), true
		}
	}

	return "", false
}

// hasHexBoundaries emulates a case-insensitive `\b` for hex runs, since Go's
// RE2 has no lookarounds: rejects matches embedded inside longer hex runs.
func hasHexBoundaries(s string, start, end int) bool {
	if start > 0 && isHexByte(s[start-1]) {
		return false
	}
	if end < len(s) && isHexByte(s[end]) {
		return false
	}
	return true
}

func isHexByte(b byte) bool {
	switch {
	case b >= '0' && b <= '9':
		return true
	case b >= 'a' && b <= 'f':
		return true
	case b >= 'A' && b <= 'F':
		return true
	}
	return false
}

// extractImageTag isolates the tag, anchoring after the last `/` so registry
// ports (host:5000/repo) are not mistaken for tags.
func extractImageTag(image string) string {
	lastSlash := strings.LastIndex(image, "/")
	rest := image
	if lastSlash >= 0 {
		rest = image[lastSlash+1:]
	}
	colon := strings.LastIndex(rest, ":")
	if colon < 0 {
		return ""
	}
	return strings.TrimSpace(rest[colon+1:])
}

func stripImageDigest(image string) string {
	if at := strings.LastIndex(image, "@"); at >= 0 {
		return image[:at]
	}
	return image
}

func stripImageTagAndDigest(image string) string {
	image = stripImageDigest(strings.TrimSpace(image))
	lastSlash := strings.LastIndex(image, "/")
	tail := image
	if lastSlash >= 0 {
		tail = image[lastSlash+1:]
	}
	if colon := strings.LastIndex(tail, ":"); colon >= 0 {
		head := ""
		if lastSlash >= 0 {
			head = image[:lastSlash+1]
		}
		return head + tail[:colon]
	}
	return image
}

func resolveImageRepoURL(image string, mappings []models.ArgocdImageRepoMapping, fallback string) string {
	repo := stripImageTagAndDigest(image)
	for _, m := range mappings {
		if m.Pattern == "" || m.RepoURL == "" {
			continue
		}
		matched, err := path.Match(m.Pattern, repo)
		if err != nil {
			continue
		}
		if matched {
			return m.RepoURL
		}
	}
	return fallback
}
