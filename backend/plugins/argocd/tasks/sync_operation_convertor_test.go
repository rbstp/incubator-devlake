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
	"time"

	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/argocd/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectEnvironment_DefaultsToTestingWhenConfigMissing(t *testing.T) {
	env := detectEnvironment(
		&models.ArgocdSyncOperation{ApplicationName: "staging-app"},
		nil,
		nil,
		nil,
	)
	assert.Equal(t, devops.TESTING, env)
}

func TestDetectEnvironment_UsesEnvPatternFallback(t *testing.T) {
	config := &models.ArgocdScopeConfig{
		EnvNamePattern: "",
	}
	env := detectEnvironment(
		&models.ArgocdSyncOperation{ApplicationName: "prod-app"},
		nil,
		config,
		nil,
	)
	assert.Equal(t, devops.PRODUCTION, env)
}

func TestDetectEnvironment_UsesRegexEnricherPriorities(t *testing.T) {
	enricher := api.NewRegexEnricher()
	assert.NoError(t, enricher.TryAdd(devops.PRODUCTION, "(?i)critical"))
	assert.NoError(t, enricher.TryAdd(devops.ENV_NAME_PATTERN, "(?i)prod"))

	config := &models.ArgocdScopeConfig{
		ProductionPattern: "(?i)critical",
		EnvNamePattern:    "(?i)prod",
	}

	env := detectEnvironment(
		&models.ArgocdSyncOperation{ApplicationName: "critical-app"},
		&models.ArgocdApplication{DestNamespace: "prod-east"},
		config,
		enricher,
	)
	assert.Equal(t, devops.PRODUCTION, env)
}

func TestIncludeSyncOperation(t *testing.T) {
	assert.True(t, includeSyncOperation(&models.ArgocdSyncOperation{ApplicationName: "any"}, nil, nil))

	config := &models.ArgocdScopeConfig{
		DeploymentPattern: "^prod",
	}
	assert.True(t, includeSyncOperation(&models.ArgocdSyncOperation{ApplicationName: "prod-app"}, config, nil))
	assert.False(t, includeSyncOperation(&models.ArgocdSyncOperation{ApplicationName: "dev-app"}, config, nil))

	config.DeploymentPattern = "(" // invalid regex should default to include
	assert.True(t, includeSyncOperation(&models.ArgocdSyncOperation{ApplicationName: "dev-app"}, config, nil))

	enricher := api.NewRegexEnricher()
	assert.NoError(t, enricher.TryAdd(devops.DEPLOYMENT, "(?i)release"))
	assert.True(t, includeSyncOperation(&models.ArgocdSyncOperation{ApplicationName: "release-app"}, config, enricher))
	assert.False(t, includeSyncOperation(&models.ArgocdSyncOperation{ApplicationName: "feature-app"}, config, enricher))
}

func TestConvertPhaseToResult(t *testing.T) {
	assert.Equal(t, devops.RESULT_SUCCESS, convertPhaseToResult("Succeeded"))
	assert.Equal(t, devops.RESULT_FAILURE, convertPhaseToResult("Failed"))
	assert.Equal(t, devops.RESULT_FAILURE, convertPhaseToResult("Error"))
	assert.Equal(t, devops.RESULT_FAILURE, convertPhaseToResult("Terminating"))
	assert.Equal(t, devops.RESULT_DEFAULT, convertPhaseToResult("Unknown"))
}

func TestConvertPhaseToStatus(t *testing.T) {
	assert.Equal(t, devops.STATUS_DONE, convertPhaseToStatus("Succeeded"))
	assert.Equal(t, devops.STATUS_DONE, convertPhaseToStatus("Failed"))
	assert.Equal(t, devops.STATUS_IN_PROGRESS, convertPhaseToStatus("Running"))
	assert.Equal(t, devops.STATUS_IN_PROGRESS, convertPhaseToStatus("Terminating"))
	assert.Equal(t, devops.STATUS_OTHER, convertPhaseToStatus("Unknown"))
}

func newTestDeployment() *devops.CICDDeployment {
	return &devops.CICDDeployment{
		Name:                "myapp:1",
		DisplayTitle:        "sync 1",
		Result:              devops.RESULT_SUCCESS,
		Status:              devops.STATUS_DONE,
		OriginalStatus:      "Succeeded",
		OriginalResult:      "Succeeded",
		Environment:         devops.PRODUCTION,
		OriginalEnvironment: "myapp",
	}
}

func TestEmitDeploymentCommits_DefaultsToRevisionBased(t *testing.T) {
	syncOp := &models.ArgocdSyncOperation{
		ApplicationName: "myapp",
		Revision:        "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1",
		RepoURL:         "https://github.com/example/manifests",
	}
	commits := emitDeploymentCommits(
		syncOp, nil, nil, newTestDeployment(),
		"argocd:depl:1", "argocd:scope:1", time.Unix(0, 0).UTC(),
	)
	require.Len(t, commits, 1)
	assert.Equal(t, "argocd:depl:1", commits[0].Id)
	assert.Equal(t, "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", commits[0].CommitSha)
	assert.Equal(t, "https://github.com/example/manifests", commits[0].RepoUrl)
}

func TestEmitDeploymentCommits_NoRevisionAndFlagOff_ReturnsNil(t *testing.T) {
	syncOp := &models.ArgocdSyncOperation{ApplicationName: "myapp"}
	commits := emitDeploymentCommits(
		syncOp, nil, nil, newTestDeployment(),
		"argocd:depl:1", "argocd:scope:1", time.Unix(0, 0).UTC(),
	)
	assert.Nil(t, commits)
}

func TestEmitDeploymentCommits_PreferImageCommit_EmitsOneRowPerParseableImage(t *testing.T) {
	syncOp := &models.ArgocdSyncOperation{
		ApplicationName: "myapp",
		Revision:        "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1",
		ContainerImages: []string{
			"registry.example.com/payments-api:abc1234",
			"registry.example.com/payments-worker:v1.2.3-def5678",
			"registry.example.com/sidecar:latest",
		},
	}
	scopeConfig := &models.ArgocdScopeConfig{
		PreferImageCommit: true,
		ImageRepoMappings: []models.ArgocdImageRepoMapping{
			{Pattern: "registry.example.com/payments-*", RepoURL: "https://github.com/example/payments"},
		},
	}
	commits := emitDeploymentCommits(
		syncOp, nil, scopeConfig, newTestDeployment(),
		"argocd:depl:1", "argocd:scope:1", time.Unix(0, 0).UTC(),
	)
	require.Len(t, commits, 2)

	bySha := map[string]*devops.CicdDeploymentCommit{}
	for _, c := range commits {
		bySha[c.CommitSha] = c
	}
	assert.Contains(t, bySha, "abc1234")
	assert.Contains(t, bySha, "def5678")
	assert.Equal(t, "https://github.com/example/payments", bySha["abc1234"].RepoUrl)
	assert.Equal(t, "https://github.com/example/payments", bySha["def5678"].RepoUrl)

	assert.NotEqual(t, commits[0].Id, commits[1].Id)
	assert.Equal(t, "argocd:depl:1", commits[0].CicdDeploymentId)
	assert.Equal(t, "argocd:depl:1", commits[1].CicdDeploymentId)
}

func TestEmitDeploymentCommits_PreferImageCommit_FallsBackToRevisionWhenNoSHA(t *testing.T) {
	syncOp := &models.ArgocdSyncOperation{
		ApplicationName: "myapp",
		Revision:        "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1",
		RepoURL:         "https://github.com/example/manifests",
		ContainerImages: []string{
			"registry.example.com/myapp:latest",
			"registry.example.com/myapp:1.21",
		},
	}
	scopeConfig := &models.ArgocdScopeConfig{
		PreferImageCommit: true,
		ImageRepoMappings: []models.ArgocdImageRepoMapping{
			{Pattern: "registry.example.com/myapp", RepoURL: "https://github.com/example/myapp-image"},
		},
	}
	commits := emitDeploymentCommits(
		syncOp, nil, scopeConfig, newTestDeployment(),
		"argocd:depl:1", "argocd:scope:1", time.Unix(0, 0).UTC(),
	)
	require.Len(t, commits, 1)
	assert.Equal(t, "argocd:depl:1", commits[0].Id)
	assert.Equal(t, "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", commits[0].CommitSha)
	assert.Equal(t, "https://github.com/example/manifests", commits[0].RepoUrl)
}

func TestEmitDeploymentCommits_PreferImageCommit_UnmappedImageGetsFallbackRepoUrl(t *testing.T) {
	syncOp := &models.ArgocdSyncOperation{
		ApplicationName: "myapp",
		Revision:        "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1",
		RepoURL:         "https://github.com/example/manifests",
		ContainerImages: []string{"registry.example.com/orphan:abc1234"},
	}
	scopeConfig := &models.ArgocdScopeConfig{
		PreferImageCommit: true,
		ImageRepoMappings: []models.ArgocdImageRepoMapping{
			{Pattern: "registry.example.com/payments-*", RepoURL: "https://github.com/example/payments"},
		},
	}
	commits := emitDeploymentCommits(
		syncOp, nil, scopeConfig, newTestDeployment(),
		"argocd:depl:1", "argocd:scope:1", time.Unix(0, 0).UTC(),
	)
	require.Len(t, commits, 1)
	assert.Equal(t, "abc1234", commits[0].CommitSha)
	assert.Equal(t, "https://github.com/example/manifests", commits[0].RepoUrl)
}

func TestEmitDeploymentCommits_PreferImageCommit_DeterministicIds(t *testing.T) {
	syncOp := &models.ArgocdSyncOperation{
		ApplicationName: "myapp",
		Revision:        "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1",
		ContainerImages: []string{
			"registry.example.com/myapp:abc1234",
			"registry.example.com/sidecar:def5678",
		},
	}
	scopeConfig := &models.ArgocdScopeConfig{PreferImageCommit: true}

	commits := emitDeploymentCommits(
		syncOp, nil, scopeConfig, newTestDeployment(),
		"argocd:depl:1", "argocd:scope:1", time.Unix(0, 0).UTC(),
	)
	require.Len(t, commits, 2)
	assert.Equal(t, "argocd:depl:1:registry.example.com/myapp:abc1234", commits[0].Id)
	assert.Equal(t, "argocd:depl:1:registry.example.com/sidecar:def5678", commits[1].Id)
}

func TestEmitDeploymentCommits_FlagOff_IgnoresImageData(t *testing.T) {
	syncOp := &models.ArgocdSyncOperation{
		ApplicationName: "myapp",
		Revision:        "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1",
		ContainerImages: []string{"registry.example.com/myapp:abc1234"},
	}
	scopeConfig := &models.ArgocdScopeConfig{PreferImageCommit: false}
	commits := emitDeploymentCommits(
		syncOp, nil, scopeConfig, newTestDeployment(),
		"argocd:depl:1", "argocd:scope:1", time.Unix(0, 0).UTC(),
	)
	require.Len(t, commits, 1)
	assert.Equal(t, "5dd95b4efd7e9b668c361bbddb8d7f1e56c32ac1", commits[0].CommitSha)
}
