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

import React from 'react';
import { Tag, Button, Space, Popconfirm } from '@douyinfe/semi-ui';
import { IconEdit, IconDelete } from '@douyinfe/semi-icons';
import { parseCapabilities } from '../utils';

// 模拟翻译函数
const defaultT = (key) => {
  const translations = {
    'ID': 'ID',
    '模型名称': '模型名称',
    '上下文长度': '上下文长度',
    '最大输出': '最大输出',
    '能力': '能力',
    '状态': '状态',
    '启用': '启用',
    '禁用': '禁用',
    '操作': '操作',
    '确认删除': '确认删除',
    '确定要删除该模型规格吗？此操作不可撤销。': '确定要删除该模型规格吗？此操作不可撤销。',
    '删除': '删除',
    '取消': '取消',
  };
  return translations[key] || key;
};

export const getModelSpecsColumnDefs = ({ t: propT, onEdit, onDelete }) => {
  const t = propT || defaultT;

  return [
    {
      title: t('ID'),
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (id) => <Tag color='grey' size='small'>{id}</Tag>,
    },
    {
      title: t('模型名称'),
      dataIndex: 'model_name',
      key: 'model_name',
      width: 200,
      render: (modelName, record) => {
        const icon = record.icon || modelName?.charAt(0) || 'N';
        return (
          <div className='flex items-center gap-2'>
            <span className='inline-flex items-center justify-center w-6 h-6 rounded-full bg-slate-100 text-slate-600 text-sm font-medium'>
              {icon?.charAt(0) || 'N'}
            </span>
            <span className='font-medium'>{modelName}</span>
          </div>
        );
      },
    },
    {
      title: t('上下文长度'),
      dataIndex: 'context_length',
      key: 'context_length',
      width: 120,
      render: (value) => (
        <span className='font-mono text-sm'>
          {value || '-'}
        </span>
      ),
    },
    {
      title: t('能力'),
      dataIndex: 'capabilities',
      key: 'capabilities',
      width: 180,
      render: (capabilities) => {
        const caps = parseCapabilities(capabilities);
        if (!caps || caps.length === 0) {
          return <span className='text-slate-400 text-sm'>-</span>;
        }
        return (
          <div className='flex flex-wrap gap-1'>
            {caps.slice(0, 3).map((cap, idx) => (
              <Tag key={idx} size='small' color='blue'>{cap}</Tag>
            ))}
            {caps.length > 3 && <Tag size='small' color='grey'>+{caps.length - 3}</Tag>}
          </div>
        );
      },
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => (
        <Tag color={status === 1 ? 'green' : 'grey'} size='small'>
          {status === 1 ? t('启用') : t('禁用')}
        </Tag>
      ),
    },
    {
      title: t('操作'),
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
          size='small'
          type='tertiary'
          theme='borderless'
          icon={<IconEdit />}
          onClick={() => onEdit(record)}
        />
        <Popconfirm
          title={t('确认删除')}
          content={t('确定要删除该模型规格吗？此操作不可撤销。')}
          onConfirm={() => onDelete(record)}
          okText={t('删除')}
          cancelText={t('取消')}
          okType='danger'
        >
          <Button
            size='small'
            type='tertiary'
            theme='borderless'
            icon={<IconDelete />}
          />
        </Popconfirm>
        </Space>
      ),
    },
  ];
};
