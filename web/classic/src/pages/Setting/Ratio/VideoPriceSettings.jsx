/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import React, { useEffect, useMemo, useState } from 'react';
import {
  Banner,
  Button,
  Card,
  Checkbox,
  Input,
  InputNumber,
  Radio,
  RadioGroup,
  Table,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import { IconCopy, IconDelete, IconPlus } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, copy, showError, showSuccess } from '../../../helpers';

const { Text } = Typography;

const OPTION_KEY = 'video_price_table.table';

// ── 数据模型与纯函数 ──
// 存储格式：Record<modelName, { resolution, has_video, price }[]>
// 解析时兼容历史 camelCase 键 "hasVideo"。

function parseTableToGroups(raw) {
  if (!raw) return [];
  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed))
    return null;

  const groups = [];
  let id = 0;
  for (const [model, variants] of Object.entries(parsed)) {
    const tiers = (Array.isArray(variants) ? variants : []).map((v) => ({
      id: id++,
      resolution: v?.resolution ?? '',
      hasVideo: v?.has_video ?? v?.hasVideo ?? false,
      price: typeof v?.price === 'number' ? v.price : 0,
    }));
    groups.push({ id: id++, model, tiers });
  }
  return groups;
}

function groupsToTable(groups) {
  const table = {};
  for (const group of groups) {
    const model = group.model?.trim();
    if (!model) continue;
    const variants = [];
    for (const tier of group.tiers) {
      const resolution = tier.resolution?.trim();
      if (!resolution) continue;
      variants.push({
        resolution,
        has_video: !!tier.hasVideo,
        price: Number(tier.price) || 0,
      });
    }
    if (variants.length > 0) table[model] = variants;
  }
  return table;
}

// 保存前校验，返回 null 表示合法
function validateGroups(groups) {
  const seenModels = new Set();
  for (const group of groups) {
    const model = group.model?.trim();
    if (!model) return 'empty_model';
    if (seenModels.has(model)) return 'duplicate_model';
    seenModels.add(model);

    const seenTiers = new Set();
    for (const tier of group.tiers) {
      const resolution = tier.resolution?.trim().toLowerCase();
      if (!resolution) return 'empty_resolution';
      if (Number(tier.price) < 0) return 'negative_price';
      const key = `${resolution}|${tier.hasVideo}`;
      if (seenTiers.has(key)) return 'duplicate_tier';
      seenTiers.add(key);
    }
  }
  return null;
}

export default function VideoPriceSettings({ options }) {
  const { t } = useTranslation();
  const [groups, setGroups] = useState([]);
  const [mode, setMode] = useState('visual');
  const [jsonText, setJsonText] = useState('');
  const [jsonError, setJsonError] = useState('');
  const [saving, setSaving] = useState(false);
  const [nextId, setNextId] = useState(0);
  // 倍率计算器（纯前端）
  const [calcPrice, setCalcPrice] = useState(null);
  const [calcMarkup, setCalcMarkup] = useState(0);
  const [calcRate, setCalcRate] = useState(1);

  const VALIDATION_MESSAGES = {
    duplicate_model: t('模型名称重复'),
    empty_model: t('模型名称不能为空'),
    empty_resolution: t('档位分辨率不能为空'),
    duplicate_tier: t(
      '档位重复：同一模型下已存在相同分辨率和视频输入标记的档位',
    ),
    negative_price: t('价格不能为负数'),
  };

  const maxIdOf = (gs) =>
    gs.reduce((acc, g) => Math.max(acc, g.id, ...g.tiers.map((x) => x.id)), -1);

  useEffect(() => {
    const parsed = parseTableToGroups(options?.[OPTION_KEY]) ?? [];
    setGroups(parsed);
    setJsonText(JSON.stringify(groupsToTable(parsed), null, 2));
    setNextId(maxIdOf(parsed) + 1);
    const rate = Number(options?.['USDExchangeRate']);
    if (rate > 0) setCalcRate(rate);
  }, [options]);

  const applyGroups = (next) => {
    setGroups(next);
    setJsonText(JSON.stringify(groupsToTable(next), null, 2));
    setJsonError('');
  };

  const syncToVisual = (text) => {
    setJsonText(text);
    const parsed = parseTableToGroups(text);
    if (parsed === null) {
      setJsonError(t('JSON 必须是对象'));
      return;
    }
    setGroups(parsed);
    setNextId(maxIdOf(parsed) + 1);
    setJsonError('');
  };

  const changeModel = (groupId, model) => {
    applyGroups(groups.map((g) => (g.id === groupId ? { ...g, model } : g)));
  };

  const changeTier = (groupId, tierId, field, value) => {
    applyGroups(
      groups.map((g) =>
        g.id === groupId
          ? {
              ...g,
              tiers: g.tiers.map((tier) =>
                tier.id === tierId ? { ...tier, [field]: value } : tier,
              ),
            }
          : g,
      ),
    );
  };

  const addModel = () => {
    const groupId = nextId;
    setNextId(groupId + 3);
    applyGroups([
      ...groups,
      {
        id: groupId,
        model: '',
        tiers: [
          {
            id: groupId + 1,
            resolution: '480p/720p',
            hasVideo: false,
            price: 0,
          },
          {
            id: groupId + 2,
            resolution: '480p/720p',
            hasVideo: true,
            price: 0,
          },
        ],
      },
    ]);
  };

  const removeModel = (groupId) => {
    applyGroups(groups.filter((g) => g.id !== groupId));
  };

  const addTier = (groupId) => {
    const tierId = nextId;
    setNextId(tierId + 1);
    applyGroups(
      groups.map((g) =>
        g.id === groupId
          ? {
              ...g,
              tiers: [
                ...g.tiers,
                { id: tierId, resolution: '', hasVideo: false, price: 0 },
              ],
            }
          : g,
      ),
    );
  };

  const removeTier = (groupId, tierId) => {
    applyGroups(
      groups.map((g) =>
        g.id === groupId
          ? { ...g, tiers: g.tiers.filter((tier) => tier.id !== tierId) }
          : g,
      ),
    );
  };

  const currentTable = useMemo(() => groupsToTable(groups), [groups]);

  const handleSave = async () => {
    if (mode === 'json' && jsonError) {
      showError(t('请先修正 JSON 错误再保存'));
      return;
    }
    const error = validateGroups(groups);
    if (error) {
      showError(VALIDATION_MESSAGES[error]);
      return;
    }
    setSaving(true);
    try {
      const res = await API.put('/api/option/', {
        key: OPTION_KEY,
        value: JSON.stringify(currentTable),
      });
      if (res.data.success) {
        showSuccess(t('保存成功'));
      } else {
        showError(res.data.message || t('保存失败'));
      }
    } catch (e) {
      showError(e.message);
    } finally {
      setSaving(false);
    }
  };

  const tierColumns = (group) => [
    {
      title: t('输出分辨率'),
      dataIndex: 'resolution',
      width: 180,
      render: (text, record) => (
        <Input
          value={text}
          placeholder='480p/720p'
          onChange={(val) => changeTier(group.id, record.id, 'resolution', val)}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('含视频输入'),
      dataIndex: 'hasVideo',
      width: 110,
      render: (val, record) => (
        <Checkbox
          checked={val}
          onChange={(e) =>
            changeTier(group.id, record.id, 'hasVideo', e.target.checked)
          }
        >
          {val ? t('是') : t('否')}
        </Checkbox>
      ),
    },
    {
      title: t('价格（元/百万tokens）'),
      dataIndex: 'price',
      width: 170,
      render: (val, record) => (
        <InputNumber
          value={val}
          min={0}
          step={0.5}
          onChange={(v) => changeTier(group.id, record.id, 'price', v ?? 0)}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('操作'),
      width: 60,
      render: (_, record) => (
        <Button
          icon={<IconDelete />}
          type='danger'
          theme='borderless'
          size='small'
          onClick={() => removeTier(group.id, record.id)}
        />
      ),
    },
  ];

  // 倍率计算器：倍率 = 目标价 × (1 + 加价率/100) ÷ 2 ÷ 汇率
  const calcRevenue =
    calcPrice != null && calcPrice > 0
      ? calcPrice * (1 + (Number(calcMarkup) || 0) / 100)
      : null;
  const calcRatio =
    calcRevenue != null && calcRate > 0 ? calcRevenue / 2 / calcRate : null;

  const calculatorCard = (
    <Card
      title={t('倍率计算器')}
      style={{ width: 300, flexShrink: 0, position: 'sticky', top: 16 }}
      headerStyle={{ padding: '12px 16px' }}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div>
          <Text size='small' type='tertiary'>
            {t('基准档目标价（元/百万tokens）')}
          </Text>
          <InputNumber
            value={calcPrice}
            min={0}
            step={1}
            placeholder='92'
            onChange={(v) => setCalcPrice(v)}
            style={{ width: '100%', marginTop: 4 }}
          />
        </div>
        <div>
          <Text size='small' type='tertiary'>
            {t('加价率（%）')}
          </Text>
          <InputNumber
            value={calcMarkup}
            step={5}
            formatter={(v) => `${v}`}
            onChange={(v) => setCalcMarkup(v ?? 0)}
            style={{ width: '100%', marginTop: 4 }}
          />
        </div>
        <div>
          <Text size='small' type='tertiary'>
            {t('汇率（1 美元兑换人民币）')}
          </Text>
          <InputNumber
            value={calcRate}
            min={0}
            step={0.1}
            onChange={(v) => setCalcRate(v ?? 1)}
            style={{ width: '100%', marginTop: 4 }}
          />
        </div>
        <div
          style={{
            borderTop: '1px solid var(--semi-color-border)',
            paddingTop: 12,
          }}
        >
          <Text size='small' type='tertiary'>
            {t('基准倍率')}
          </Text>
          <div style={{ fontSize: 28, fontWeight: 600, lineHeight: 1.3 }}>
            {calcRatio != null ? calcRatio.toFixed(2) : '-'}
          </div>
          {calcRevenue != null && (
            <Text size='small' type='tertiary'>
              {t('每百万 tokens 实收')} {calcRevenue.toFixed(2)} {t('元')}
            </Text>
          )}
        </div>
        <Text size='small' type='tertiary'>
          {t('把该值填入「模型定价设置」中该模型的倍率即可。')}
        </Text>
      </div>
    </Card>
  );

  return (
    <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
      <div style={{ flex: 1, minWidth: 0, maxWidth: 900 }}>
        <Banner
          type='info'
          description={
            <>
              <div>
                {t(
                  '按模型配置视频生成价格：先添加模型，再按输出分辨率和是否含视频输入添加价格档位。',
                )}
              </div>
              <div style={{ marginTop: 4 }}>
                {t(
                  '档位价格只决定相对倍率：价格最低的无视频档为基准价（按模型倍率计费），其余档位按价格比例浮动。分辨率支持 480p/720p 组合写法。',
                )}
              </div>
            </>
          }
          style={{ marginBottom: 16 }}
        />

        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 12,
            flexWrap: 'wrap',
            gap: 8,
          }}
        >
          {mode === 'visual' ? (
            <Button icon={<IconPlus />} onClick={addModel}>
              {t('添加模型')}
            </Button>
          ) : (
            <span />
          )}
          <RadioGroup
            type='button'
            size='small'
            value={mode}
            onChange={(e) => setMode(e.target.value)}
          >
            <Radio value='visual'>{t('可视化')}</Radio>
            <Radio value='json'>JSON</Radio>
          </RadioGroup>
        </div>

        {mode === 'visual' ? (
          groups.length === 0 ? (
            <Card>
              <Text type='tertiary'>{t('未配置视频模型')}</Text>
            </Card>
          ) : (
            groups.map((group) => (
              <Card
                key={group.id}
                style={{ marginBottom: 16 }}
                headerStyle={{ padding: '12px 16px' }}
                title={
                  <div
                    style={{ display: 'flex', alignItems: 'center', gap: 8 }}
                  >
                    <Input
                      value={group.model}
                      placeholder={t('模型名称，如 doubao-seedance-2-0')}
                      onChange={(val) => changeModel(group.id, val)}
                      style={{ width: 320, fontFamily: 'monospace' }}
                    />
                    <Button
                      icon={<IconDelete />}
                      type='danger'
                      theme='borderless'
                      size='small'
                      title={t('删除模型')}
                      onClick={() => removeModel(group.id)}
                    />
                  </div>
                }
              >
                <Table
                  dataSource={group.tiers}
                  columns={tierColumns(group)}
                  pagination={false}
                  size='small'
                  rowKey='id'
                  empty={<Text type='tertiary'>{t('暂无档位')}</Text>}
                />
                <Button
                  icon={<IconPlus />}
                  size='small'
                  style={{ marginTop: 8 }}
                  onClick={() => addTier(group.id)}
                >
                  {t('添加档位')}
                </Button>
              </Card>
            ))
          )
        ) : (
          <>
            <TextArea
              value={jsonText}
              onChange={syncToVisual}
              autosize={{ minRows: 10, maxRows: 30 }}
              style={{ fontFamily: 'monospace', fontSize: 13 }}
            />
            {jsonError && (
              <Text
                type='danger'
                size='small'
                style={{ display: 'block', marginTop: 4 }}
              >
                {jsonError}
              </Text>
            )}
            <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
              <Button
                icon={<IconCopy />}
                size='small'
                theme='borderless'
                onClick={() => {
                  copy(jsonText, t('JSON'));
                }}
              >
                {t('复制')}
              </Button>
            </div>
          </>
        )}

        <div
          style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16 }}
        >
          <Button
            theme='solid'
            type='primary'
            loading={saving}
            disabled={mode === 'json' && !!jsonError}
            onClick={handleSave}
          >
            {t('保存视频模型价格')}
          </Button>
        </div>
      </div>
      {calculatorCard}
    </div>
  );
}
