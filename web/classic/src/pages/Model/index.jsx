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
import React, { useState, useEffect, useMemo, Suspense } from 'react';
import { useTranslation } from 'react-i18next';
import { Tabs, TabPane } from '@douyinfe/semi-ui';
import ModelsTable from '../../components/table/models';

// Sidecar: 最小侵入式集成模型规格配置
import * as modelSpecsSidecar from '../../sidecar/model-specs/register';

const ModelPage = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('metadata');

  // 从 URL hash 读取标签
  useEffect(() => {
    const hash = window.location.hash.slice(1);
    if (hash === 'specs' || hash === 'metadata') {
      setActiveTab(hash);
    }
  }, []);

  // 更新 URL hash
  const handleTabChange = (key) => {
    setActiveTab(key);
    window.location.hash = key;
  };

  // 构建标签页列表
  const tabs = useMemo(() => {
    const baseTabs = [
      {
        key: 'metadata',
        label: t('模型元数据'),
        component: <ModelsTable />,
      },
    ];

    // 如果 sidecar 组件可用，添加标签页
    if (modelSpecsSidecar?.ModelSpecs && modelSpecsSidecar?.getModelSpecsTab) {
      try {
        const config = modelSpecsSidecar.getModelSpecsTab();
        baseTabs.push({
          key: config.key,
          label: t(config.label),
          component: <modelSpecsSidecar.ModelSpecs t={t} />,
        });
      } catch (e) {
        console.warn('Failed to add model specs tab:', e);
      }
    }

    return baseTabs;
  }, [t]);

  return (
    <div className='mt-[60px] px-2'>
      <Tabs
        activeKey={activeTab}
        onChange={handleTabChange}
        type='line'
        className='mb-4'
      >
        {tabs.map((tab) => (
          <TabPane
            key={tab.key}
            itemKey={tab.key}
            tab={tab.label}
          >
            <Suspense fallback={<div>加载中...</div>}>
              {tab.component}
            </Suspense>
          </TabPane>
        ))}
      </Tabs>
    </div>
  );
};

export default ModelPage;
