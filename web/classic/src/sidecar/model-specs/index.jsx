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

import React, { useState, useEffect, useCallback } from 'react';
import { Button, Input, Space, Typography } from '@douyinfe/semi-ui';
import { IconPlus, IconSearch } from '@douyinfe/semi-icons';
import { API, showError, showSuccess } from '../../helpers';
import { createCardProPagination } from '../../helpers/utils';
import CardPro from '../../components/common/ui/CardPro';
import ModelSpecsTable from './components/ModelSpecsTable';
import ModelSpecEditModal from './components/modals/ModelSpecEditModal';

const PAGE_SIZE = 20;

// 模拟翻译函数
const defaultT = (key) => {
  const translations = {
    '模型规格配置': '模型规格配置',
    '管理模型的能力配置和规格信息': '管理模型的能力配置和规格信息',
    '搜索模型名称...': '搜索模型名称...',
    '添加模型规格': '添加模型规格',
    '获取数据失败': '获取数据失败',
    '删除成功': '删除成功',
    '删除失败': '删除失败',
    '更新成功': '更新成功',
    '创建成功': '创建成功',
    '操作失败': '操作失败',
    'ID': 'ID',
    '模型名称': '模型名称',
    '描述': '描述',
    '上下文长度': '上下文长度',
    '最大输出': '最大输出',
    '能力': '能力',
    '状态': '状态',
    '启用': '启用',
    '禁用': '禁用',
    '操作': '操作',
    '显示操作项': '显示操作项',
    '隐藏操作项': '隐藏操作项',
  };
  return translations[key] || key;
};

const ModelSpecs = ({ t: propT }) => {
  const t = propT || defaultT;
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [data, setData] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [searchText, setSearchText] = useState('');

  const [modalVisible, setModalVisible] = useState(false);
  const [editingSpec, setEditingSpec] = useState(null);

  // 加载数据
  const loadData = useCallback(async (page = 1, search = '') => {
    setLoading(true);
    try {
      const params = {
        p: page,
        page_size: PAGE_SIZE,
      };
      if (search) {
        params.keyword = search;
      }

      const res = await API.get('/api/model-catalog/admin/specs', { params });
      if (res.data?.success) {
        setData(res.data.data?.items || []);
        setTotal(res.data.data?.total || 0);
        setCurrentPage(page);
      } else {
        showError(res.data?.message || t('获取数据失败'));
      }
    } catch (error) {
      showError(error.message || t('获取数据失败'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  // 初始加载
  useEffect(() => {
    loadData(1, searchText);
  }, []);

  // 搜索
  const handleSearch = useCallback(() => {
    loadData(1, searchText);
  }, [searchText, loadData]);

  // 分页变化
  const handlePageChange = useCallback((page) => {
    loadData(page, searchText);
  }, [searchText, loadData]);

  // 创建
  const handleCreate = useCallback(() => {
    setEditingSpec(null);
    setModalVisible(true);
  }, []);

  // 编辑
  const handleEdit = useCallback((spec) => {
    setEditingSpec(spec);
    setModalVisible(true);
  }, []);

  // 删除
  const handleDelete = useCallback(async (spec) => {
    try {
      const res = await API.delete(`/api/model-catalog/admin/specs/${spec.id}`);
      if (res.data?.success) {
        showSuccess(t('删除成功'));
        loadData(currentPage, searchText);
      } else {
        showError(res.data?.message || t('删除失败'));
      }
    } catch (error) {
      showError(error.message || t('删除失败'));
    }
  }, [currentPage, searchText, loadData, t]);

  // 提交表单
  const handleSubmit = useCallback(async (values, isEdit) => {
    setSubmitting(true);
    try {
      let res;
      if (isEdit) {
        res = await API.post('/api/model-catalog/admin/specs', values);
      } else {
        res = await API.post('/api/model-catalog/admin/specs', values);
      }

      if (res.data?.success) {
        showSuccess(isEdit ? t('更新成功') : t('创建成功'));
        setModalVisible(false);
        loadData(currentPage, searchText);
      } else {
        showError(res.data?.message || t('操作失败'));
      }
    } catch (error) {
      showError(error.message || t('操作失败'));
    } finally {
      setSubmitting(false);
    }
  }, [currentPage, searchText, loadData, t]);

  // 描述区域
  const descriptionArea = (
    <div>
      <Typography.Title heading={5} style={{ margin: 0 }}>
        {t('模型规格配置')}
      </Typography.Title>
      <Typography.Text type='tertiary' style={{ marginTop: 4 }}>
        {t('管理模型的能力配置和规格信息')}
      </Typography.Text>
    </div>
  );

  // 操作区域
  const actionsArea = (
    <div className='flex flex-col md:flex-row justify-between items-center gap-2 w-full'>
      <div className='w-full md:w-auto order-2 md:order-1'>
        <Input
          placeholder={t('搜索模型名称...')}
          value={searchText}
          onChange={setSearchText}
          onEnterPress={handleSearch}
          prefix={<IconSearch />}
          style={{ width: 300 }}
          showClear
        />
      </div>
      <div className='w-full md:w-auto order-1 md:order-2'>
        <Button
          theme='solid'
          type='primary'
          icon={<IconPlus />}
          onClick={handleCreate}
        >
          {t('添加模型规格')}
        </Button>
      </div>
    </div>
  );

  return (
    <>
      <CardPro
        type='type1'
        descriptionArea={descriptionArea}
        actionsArea={actionsArea}
        paginationArea={createCardProPagination({
          currentPage,
          pageSize: PAGE_SIZE,
          total,
          onPageChange: handlePageChange,
          t,
        })}
        t={t}
      >
        <ModelSpecsTable
          dataSource={data}
          loading={loading}
          onEdit={handleEdit}
          onDelete={handleDelete}
          t={t}
        />
      </CardPro>

      <ModelSpecEditModal
        visible={modalVisible}
        editingSpec={editingSpec}
        onCancel={() => setModalVisible(false)}
        onOk={handleSubmit}
        loading={submitting}
        t={t}
      />
    </>
  );
};

export default ModelSpecs;
