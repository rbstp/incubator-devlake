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

var _ plugin.MigrationScript = (*addIssueFixVerion)(nil)

type addIssueFixVerion struct{}

type issue20250710 struct {
	FixVersions string `gorm:"type:text"`
}

func (issue20250710) TableName() string {
	return "issues"
}

func (script *addIssueFixVerion) Up(basicRes context.BasicRes) errors.Error {
	return basicRes.GetDal().AutoMigrate(new(issue20250710))
}

func (*addIssueFixVerion) Version() uint64 {
	return 20250710140418
}

func (*addIssueFixVerion) Name() string {
	return "add missing issues.fix_versions"
}
