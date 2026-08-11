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

import React, { useState, useEffect, useRef, useMemo } from 'react';
import {
  SideSheet,
  Form,
  Input,
  InputNumber,
  Checkbox,
  Switch,
  Typography,
  TextArea,
  Button,
  Space,
  Spin,
  Tag,
} from '@douyinfe/semi-ui';
import { Save, X } from 'lucide-react';
import { CAPABILITY_OPTIONS, parseCapabilities, serializeCapabilities } from '../../utils';
import { API } from '../../../../helpers';

const { Title } = Typography;

// 模拟翻译函数
const defaultT = (key) => {
  const translations = {
    '编辑模型规格': '编辑模型规格',
    '创建模型规格': '创建模型规格',
    '基本信息': '基本信息',
    '模型名称': '模型名称',
    '请选择模型': '请选择模型',
    '规格信息': '规格信息',
    '上下文长度': '上下文长度',
    '最大输出 Tokens': '最大输出 Tokens',
    '能力配置': '能力配置',
    '元数据': '元数据',
    '发布日期': '发布日期',
    '知识截止': '知识截止',
    '参数量': '参数量',
    '状态': '状态',
    '启用': '启用',
    '禁用': '禁用',
    '启用或禁用该模型规格': '启用或禁用该模型规格',
    '确认': '确认',
    '取消': '取消',
  };
  return translations[key] || key;
};

const ModelSpecEditModal = ({
  visible,
  editingSpec,
  onCancel,
  onOk,
  loading,
  t: propT,
}) => {
  const t = propT || defaultT;
  const formApiRef = useRef(null);
  const isEdit = editingSpec && editingSpec.id !== undefined;
  const placement = useMemo(() => (isEdit ? 'right' : 'right'), [isEdit]);
  const [capCount, setCapCount] = useState(
    editingSpec?.capabilities ? parseCapabilities(editingSpec.capabilities).length : 0
  );
  const [modelOptions, setModelOptions] = useState([]);
  const capabilityOptions = useMemo(
    () =>
      CAPABILITY_OPTIONS.map((option) => ({
        ...option,
        label: t(option.label),
      })),
    [t]
  );

  // 弹窗打开时加载已有模型列表
  useEffect(() => {
    if (visible) {
      API.get('/api/model-catalog').then((res) => {
        if (res.data?.success && Array.isArray(res.data.data)) {
          const names = [...new Set(res.data.data.map((m) => m.model_name))].sort();
          setModelOptions(names.map((n) => ({ value: n, label: n })));
        }
      });
    }
  }, [visible]);

  const getInitValues = () => ({
    model_name: editingSpec?.model_name || '',
    context_length: editingSpec?.context_length ?? 0,
    capabilities: editingSpec?.capabilities ? parseCapabilities(editingSpec.capabilities) : [],
    release_date: editingSpec?.release_date || '',
    parameter_count: editingSpec?.parameter_count || '',
    status: editingSpec ? editingSpec.status === 1 : true,
  });

  useEffect(() => {
    if (visible && formApiRef.current) {
      formApiRef.current.setValues(getInitValues());
      const caps = editingSpec?.capabilities ? parseCapabilities(editingSpec.capabilities) : [];
      setCapCount(caps.length);
    }
  }, [visible, editingSpec?.id]);

  const handleSubmit = async (values) => {
    const submitData = {
      ...values,
      id: isEdit ? editingSpec.id : undefined,
      // 编辑模式下 model_name 字段 disabled，表单不会提交该值，需手动补上
      model_name: isEdit ? editingSpec.model_name : values.model_name,
      // 后端 ShouldBindJSON 期望 []string，直接传数组
      capabilities: values.capabilities || [],
      // 后端 ContextLength 为 int，空字符串会导致 JSON 解析失败
      context_length: Number(values.context_length) || 0,
      status: values.status ? 1 : 0,
    };
    onOk(submitData, isEdit);
  };

  const getFormApi = (api) => {
    formApiRef.current = api;
    if (api && api.watch) {
      api.watch((field, value) => {
        if (field === 'capabilities') {
          if (value.length > 5) {
            api.setValue('capabilities', value.slice(0, 5));
            setCapCount(5);
          } else {
            setCapCount(value.length);
          }
        }
      });
    }
  };

  return (
    <SideSheet
      placement={placement}
      title={
        <Space>
          {isEdit ? (
            <Tag color='blue' shape='circle'>
              {t('更新')}
            </Tag>
          ) : (
            <Tag color='green' shape='circle'>
              {t('新建')}
            </Tag>
          )}
          <Title heading={4} className='m-0'>
            {isEdit ? t('编辑模型规格') : t('创建模型规格')}
          </Title>
        </Space>
      }
      bodyStyle={{ padding: '0' }}
      visible={visible}
      width={600}
      footer={
        <div className='flex justify-end'>
          <Space>
            <Button
              theme='solid'
              className='!rounded-lg'
              onClick={() => formApiRef.current?.submitForm()}
              icon={<Save size={16} />}
              loading={loading}
            >
              {t('确认')}
            </Button>
            <Button
              theme='light'
              className='!rounded-lg'
              type='primary'
              onClick={onCancel}
              icon={<X size={16} />}
            >
              {t('取消')}
            </Button>
          </Space>
        </div>
      }
      closeIcon={null}
      onCancel={onCancel}
    >
      <Spin spinning={loading}>
        <Form
          key={isEdit ? 'edit' : 'new'}
          initValues={getInitValues()}
          getFormApi={getFormApi}
          onSubmit={handleSubmit}
          className='p-4'
        >
          <div className='space-y-6'>
            {/* 基本信息 */}
            <div className='space-y-4'>
              <Title heading={6} className='mb-0'>{t('基本信息')}</Title>

              <Form.Select
                field='model_name'
                label={t('模型名称')}
                placeholder={t('请选择模型')}
                disabled={isEdit}
                filter
                showSearchClear
                style={{ width: '100%' }}
                optionList={modelOptions}
                rules={[{ required: true, message: t('请选择模型') }]}
              />
            </div>

            {/* 规格信息 */}
            <div className='space-y-4'>
              <Title heading={6} className='mb-0'>{t('规格信息')}</Title>

              <div className='grid grid-cols-2 gap-4'>
                <Form.InputNumber
                  field='context_length'
                  label={t('上下文长度')}
                  placeholder='128000'
                  style={{ width: '100%' }}
                />

                <Form.Input
                  field='parameter_count'
                  label={t('参数量')}
                  placeholder='1.76T'
                  style={{ width: '100%' }}
                />
              </div>

              <Form.CheckboxGroup
                field='capabilities'
                label={t('能力配置')}
                labelExtra={<span className='text-xs text-gray-400'>（{capCount}/5）</span>}
                direction='horizontal'
                style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 8 }}
                options={capabilityOptions}
              />
            </div>

            {/* 元数据 */}
            <div className='space-y-4'>
              <Title heading={6} className='mb-0'>{t('元数据')}</Title>

              <div className='grid grid-cols-1 gap-4'>
                <Form.Input
                  field='release_date'
                  label={t('发布日期')}
                  placeholder='2024-01'
                />
              </div>

              <div className='flex items-center justify-between rounded-lg border p-4'>
                <div className='space-y-0.5'>
                  <Typography.Text strong>{t('状态')}</Typography.Text>
                  <Typography.Text type='tertiary' size='small'>
                    {t('启用或禁用该模型规格')}
                  </Typography.Text>
                </div>
                <Form.Switch field='status' noLabel />
              </div>
            </div>
          </div>
        </Form>
      </Spin>
    </SideSheet>
  );
};

export default ModelSpecEditModal;
