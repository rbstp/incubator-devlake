/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { CaretRightOutlined, PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { theme, Collapse, Tag, Input, Switch, Button, Space } from 'antd';

import { HelpTooltip } from '@/components';

type ImageRepoMapping = { pattern: string; repoURL: string };

interface Props {
  entities: string[];
  transformation: any;
  setTransformation: React.Dispatch<React.SetStateAction<any>>;
}

export const ArgoCDTransformation = ({ entities, transformation, setTransformation }: Props) => {
  const { token } = theme.useToken();

  const panelStyle: React.CSSProperties = {
    marginBottom: 24,
    background: token.colorFillAlter,
    borderRadius: token.borderRadiusLG,
    border: 'none',
  };

  return (
    <Collapse
      bordered={false}
      defaultActiveKey={['CICD']}
      expandIcon={({ isActive }) => <CaretRightOutlined rotate={isActive ? 90 : 0} rev="" />}
      style={{ background: token.colorBgContainer }}
      size="large"
      items={renderCollapseItems({
        entities,
        panelStyle,
        transformation,
        onChangeTransformation: setTransformation,
      })}
    />
  );
};

const renderCollapseItems = ({
  entities,
  panelStyle,
  transformation,
  onChangeTransformation,
}: {
  entities: string[];
  panelStyle: React.CSSProperties;
  transformation: any;
  onChangeTransformation: any;
}) =>
  [
    {
      key: 'CICD',
      label: 'CI/CD',
      style: panelStyle,
      children: (
        <>
          <h3 style={{ marginBottom: 16 }}>
            <span>Deployment</span>
            <Tag style={{ marginLeft: 4 }} color="blue">
              DORA
            </Tag>
          </h3>
          <p style={{ marginBottom: 16 }}>
            Use Regular Expressions to define how DevLake identifies deployments and production environments from ArgoCD
            sync operations to measure DORA metrics.
          </p>
          <div style={{ marginTop: 16 }}>
            <strong>Environment Detection</strong>
          </div>
          <div style={{ margin: '8px 0', paddingLeft: 28 }}>
            <span>An ArgoCD sync operation is a 'Production Deployment' when the application name matches</span>
            <Input
              style={{ width: 200, margin: '0 8px' }}
              placeholder="(?i)prod(.*)"
              value={transformation.envNamePattern ?? '(?i)prod(.*)'}
              onChange={(e) =>
                onChangeTransformation({
                  ...transformation,
                  envNamePattern: e.target.value,
                })
              }
            />
            <i style={{ color: '#E34040' }}>*</i>
            <HelpTooltip content="Use regex to match application names. Default pattern matches any name containing 'prod' (case-insensitive)." />
          </div>
          <div style={{ marginTop: 16, marginBottom: 8 }}>
            <strong>Optional Filters</strong>
          </div>
          <div style={{ margin: '8px 0', paddingLeft: 28 }}>
            <span>Deployment Pattern (filter sync operations by application name)</span>
            <Input
              style={{ width: 200, margin: '0 8px' }}
              placeholder=".*"
              value={transformation.deploymentPattern ?? ''}
              onChange={(e) =>
                onChangeTransformation({
                  ...transformation,
                  deploymentPattern: e.target.value,
                })
              }
            />
            <HelpTooltip content="Optional: Use regex to include only specific applications. Leave empty to include all." />
          </div>
          <div style={{ margin: '8px 0', paddingLeft: 28 }}>
            <span>Production Pattern (additional pattern for production detection)</span>
            <Input
              style={{ width: 200, margin: '0 8px' }}
              placeholder=""
              value={transformation.productionPattern ?? ''}
              onChange={(e) =>
                onChangeTransformation({
                  ...transformation,
                  productionPattern: e.target.value,
                })
              }
            />
            <HelpTooltip content="Optional: Additional regex pattern to identify production deployments." />
          </div>
          <div style={{ marginTop: 24, marginBottom: 8 }}>
            <strong>Source Commit Resolution</strong>
          </div>
          <div style={{ margin: '8px 0', paddingLeft: 28 }}>
            <Space align="center">
              <Switch
                checked={!!transformation.preferImageCommit}
                onChange={(checked) =>
                  onChangeTransformation({
                    ...transformation,
                    preferImageCommit: checked,
                  })
                }
              />
              <span>Derive deployment commit SHA from deployed image tag</span>
              <HelpTooltip content="When ArgoCD syncs from a separate manifests repo, the synced Revision is the manifests-repo SHA, not the source-code SHA. Enable this to parse a git SHA out of each deployed image's tag (e.g. :v1.2.3-abc1234) and emit one cicd_deployment_commit per parseable image. Falls back to the Revision-based row when no image yields a SHA." />
            </Space>
          </div>
          {transformation.preferImageCommit && (
            <div style={{ margin: '12px 0 0 28px' }}>
              <div style={{ marginBottom: 4 }}>
                <span>Image Repo Mappings</span>
                <HelpTooltip content="Each mapping pairs an image-ref glob (path.Match syntax: '*' does not span '/') with the source repo URL whose commit SHA the image tag encodes. First match wins. Images that don't match any mapping fall back to the manifests repo URL." />
              </div>
              {(transformation.imageRepoMappings ?? []).map((m: ImageRepoMapping, idx: number) => (
                <div key={idx} style={{ display: 'flex', gap: 8, marginBottom: 8, alignItems: 'center' }}>
                  <Input
                    style={{ flex: 1 }}
                    placeholder="ghcr.io/your-org/myapp"
                    value={m.pattern ?? ''}
                    onChange={(e) => {
                      const next = [...(transformation.imageRepoMappings ?? [])];
                      next[idx] = { ...next[idx], pattern: e.target.value };
                      onChangeTransformation({ ...transformation, imageRepoMappings: next });
                    }}
                  />
                  <Input
                    style={{ flex: 2 }}
                    placeholder="https://github.com/your-org/myapp"
                    value={m.repoURL ?? ''}
                    onChange={(e) => {
                      const next = [...(transformation.imageRepoMappings ?? [])];
                      next[idx] = { ...next[idx], repoURL: e.target.value };
                      onChangeTransformation({ ...transformation, imageRepoMappings: next });
                    }}
                  />
                  <Button
                    type="text"
                    icon={<MinusCircleOutlined />}
                    onClick={() => {
                      const next = (transformation.imageRepoMappings ?? []).filter(
                        (_: ImageRepoMapping, i: number) => i !== idx,
                      );
                      onChangeTransformation({ ...transformation, imageRepoMappings: next });
                    }}
                  />
                </div>
              ))}
              <Button
                type="dashed"
                icon={<PlusOutlined />}
                onClick={() =>
                  onChangeTransformation({
                    ...transformation,
                    imageRepoMappings: [...(transformation.imageRepoMappings ?? []), { pattern: '', repoURL: '' }],
                  })
                }
              >
                Add mapping
              </Button>
            </div>
          )}
          <div style={{ marginTop: 16, padding: '8px 12px', background: '#f0f7ff', borderRadius: '4px' }}>
            <strong>Note:</strong> ArgoCD limits deployment history to the last 10 sync operations by default
            (controlled by <code>revisionHistoryLimit</code>). Consider increasing this value in your ArgoCD application
            settings for better historical metrics.
          </div>
        </>
      ),
    },
  ].filter((it) => entities.includes(it.key));
