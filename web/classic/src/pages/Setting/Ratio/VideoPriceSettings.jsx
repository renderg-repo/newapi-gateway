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

function parseTable(raw) {
  if (!raw) return {};
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed;
    }
  } catch {}
  return {};
}

function tableToRows(table) {
  const rows = [];
  let id = 0;
  for (const [model, variants] of Object.entries(table)) {
    for (const v of variants) {
      rows.push({
        id,
        model,
        resolution: v.resolution,
        hasVideo: v.hasVideo,
        price: v.price,
      });
      id += 1;
    }
  }
  return rows;
}

function rowsToTable(rows) {
  const table = {};
  for (const row of rows) {
    if (!row.model?.trim()) continue;
    const model = row.model.trim();
    if (!table[model]) table[model] = [];
    table[model].push({
      resolution: row.resolution?.trim() || '480p',
      hasVideo: !!row.hasVideo,
      price: Number(row.price) || 0,
    });
  }
  return table;
}

// 按 (model, resolution, hasVideo) 排序，相近档位相邻展示
function sortRows(rows) {
  return [...rows].sort((a, b) => {
    if (a.model !== b.model) return a.model.localeCompare(b.model);
    if (a.resolution !== b.resolution)
      return a.resolution.localeCompare(b.resolution);
    return (a.hasVideo ? 1 : 0) - (b.hasVideo ? 1 : 0);
  });
}

export default function VideoPriceSettings({ options }) {
  const { t } = useTranslation();
  const [rows, setRows] = useState([]);
  const [mode, setMode] = useState('visual');
  const [jsonText, setJsonText] = useState('');
  const [jsonError, setJsonError] = useState('');
  const [saving, setSaving] = useState(false);
  const [nextId, setNextId] = useState(0);

  useEffect(() => {
    const table = parseTable(options?.[OPTION_KEY]);
    const initial = sortRows(tableToRows(table));
    setRows(initial);
    setJsonText(JSON.stringify(table, null, 2));
    setNextId(initial.length);
  }, [options]);

  const syncToJson = (nextRows) => {
    const next = sortRows(nextRows);
    setRows(next);
    setJsonText(JSON.stringify(rowsToTable(next), null, 2));
    setJsonError('');
  };

  const syncToVisual = (text) => {
    setJsonText(text);
    try {
      const parsed = JSON.parse(text);
      if (typeof parsed !== 'object' || Array.isArray(parsed) || parsed === null) {
        setJsonError(t('JSON 必须是对象'));
        return;
      }
      const nextRows = tableToRows(parsed);
      setRows(sortRows(nextRows));
      setNextId(nextRows.length);
      setJsonError('');
    } catch (e) {
      setJsonError(e.message);
    }
  };

  const updateRow = (id, field, value) => {
    syncToJson(rows.map((r) => (r.id === id ? { ...r, [field]: value } : r)));
  };

  const addRow = () => {
    syncToJson([
      ...rows,
      { id: nextId, model: '', resolution: '480p', hasVideo: false, price: 0 },
    ]);
    setNextId((prev) => prev + 1);
  };

  const removeRow = (id) => {
    syncToJson(rows.filter((r) => r.id !== id));
  };

  const currentTable = useMemo(() => rowsToTable(rows), [rows]);

  const handleSave = async () => {
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

  const columns = [
    {
      title: t('模型 ID'),
      dataIndex: 'model',
      width: 260,
      render: (text, record) => (
        <Input
          value={text}
          placeholder='doubao-seedance-2-0-260128'
          onChange={(val) => updateRow(record.id, 'model', val)}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('输出分辨率'),
      dataIndex: 'resolution',
      width: 140,
      render: (text, record) => (
        <Input
          value={text}
          placeholder='1080p / 720p / 480p'
          onChange={(val) => updateRow(record.id, 'resolution', val)}
          style={{ width: '100%' }}
        />
      ),
    },
    {
      title: t('含视频输入'),
      dataIndex: 'hasVideo',
      width: 120,
      render: (val, record) => (
        <Checkbox
          checked={val}
          onChange={(e) => updateRow(record.id, 'hasVideo', e.target.checked)}
        >
          {val ? t('是') : t('否')}
        </Checkbox>
      ),
    },
    {
      title: t('价格（元/百万tokens）'),
      dataIndex: 'price',
      width: 180,
      render: (val, record) => (
        <InputNumber
          value={val}
          min={0}
          step={0.5}
          onChange={(v) => updateRow(record.id, 'price', v ?? 0)}
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
          onClick={() => removeRow(record.id)}
        />
      ),
    },
  ];

  return (
    <div style={{ maxWidth: 900 }}>
      <Banner
        type='info'
        description={
          <>
            <div>{t('配置视频模型在不同输出分辨率和视频输入场景下的单价（元/百万tokens）。模型 ID、分辨率、是否含视频输入均可自由增删配置。')}</div>
            <div style={{ marginTop: 4 }}>
              {t('计费时取「该档位价格 / 不含视频的最低基准价」作为倍率。')}
            </div>
          </>
        }
        style={{ marginBottom: 16 }}
      />

      <RadioGroup
        type='button'
        size='small'
        value={mode}
        onChange={(e) => setMode(e.target.value)}
        style={{ marginBottom: 12 }}
      >
        <Radio value='visual'>{t('可视化')}</Radio>
        <Radio value='json'>JSON</Radio>
      </RadioGroup>

      {mode === 'visual' ? (
        <>
          <Table
            dataSource={rows}
            columns={columns}
            pagination={false}
            size='small'
            rowKey='id'
          />
          <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
            <Button icon={<IconPlus />} onClick={addRow}>
              {t('添加档位')}
            </Button>
          </div>
        </>
      ) : (
        <>
          <TextArea
            value={jsonText}
            onChange={syncToVisual}
            autosize={{ minRows: 10, maxRows: 30 }}
            style={{ fontFamily: 'monospace', fontSize: 13 }}
          />
          {jsonError && (
            <Text type='danger' size='small' style={{ display: 'block', marginTop: 4 }}>
              {jsonError}
            </Text>
          )}
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <Button
              icon={<IconCopy />}
              size='small'
              theme='borderless'
              onClick={() => { copy(jsonText, t('JSON')); }}
            >
              {t('复制')}
            </Button>
          </div>
        </>
      )}

      <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16 }}>
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
  );
}