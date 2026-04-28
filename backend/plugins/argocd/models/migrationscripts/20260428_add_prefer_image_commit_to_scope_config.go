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

package migrationscripts

import (
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
)

var _ plugin.MigrationScript = (*addPreferImageCommitToScopeConfig)(nil)

type addPreferImageCommitToScopeConfig struct{}

// addPreferImageCommitScopeConfigArchived is a snapshot of ArgocdScopeConfig
// used solely for this migration so the live model can evolve independently.
type addPreferImageCommitScopeConfigArchived struct {
	Id                uint64 `gorm:"primaryKey"`
	PreferImageCommit bool   `gorm:"column:prefer_image_commit;default:false"`
	ImageRepoMappings string `gorm:"column:image_repo_mappings;type:json"`
}

func (addPreferImageCommitScopeConfigArchived) TableName() string {
	return "_tool_argocd_scope_configs"
}

func (m *addPreferImageCommitToScopeConfig) Up(basicRes context.BasicRes) errors.Error {
	db := basicRes.GetDal()
	return db.AutoMigrate(&addPreferImageCommitScopeConfigArchived{})
}

func (*addPreferImageCommitToScopeConfig) Version() uint64 {
	return 20260428000000
}

func (*addPreferImageCommitToScopeConfig) Name() string {
	return "argocd add prefer_image_commit and image_repo_mappings to scope config"
}
