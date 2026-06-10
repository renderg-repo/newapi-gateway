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

// Sidecar 模型规格功能注册器 - 最小侵入式集成
import React from 'react';
import ModelSpecs from './index';

// 导出标签页配置
export const ModelSpecsTab = {
  key: 'specs',
  label: '模型规格',
  component: ModelSpecs,
};

// 导出组件，供主项目按需引入
export { ModelSpecs };

// 导出默认标签页列表
export const modelSpecsTabs = [ModelSpecsTab];

// 兼容函数 - 获取标签页配置
export function getModelSpecsTab() {
  return ModelSpecsTab;
}
