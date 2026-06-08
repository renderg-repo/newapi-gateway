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

// 模型规格能力配置 - Sidecar版本，不依赖主项目utils
export const CAPABILITY_OPTIONS = [
  { value: 'vision', label: '视觉' },
  { value: 'reasoning', label: '推理' },
  { value: 'code', label: '代码' },
  { value: 'tool-use', label: '工具调用' },
  { value: 'audio', label: '音频' },
  { value: 'image', label: '图像' },
  { value: 'video', label: '视频' },
  { value: 'search', label: '搜索' },
  { value: 'long-context', label: '长上下文' },
];

// 解析模型规格能力
export function parseCapabilities(capabilitiesStr) {
  if (!capabilitiesStr) return [];
  try {
    const parsed = JSON.parse(capabilitiesStr);
    if (Array.isArray(parsed)) return parsed;
  } catch {
    // 如果不是数组，尝试以逗号分隔的字符串
    if (typeof capabilitiesStr === 'string' && capabilitiesStr.includes(',')) {
      return capabilitiesStr.split(',').map(s => s.trim()).filter(Boolean);
    }
  }
  return [];
}

// 序列化模型规格能力
export function serializeCapabilities(capabilities) {
  if (!Array.isArray(capabilities)) return '[]';
  try {
    return JSON.stringify(capabilities);
  } catch {
    return '[]';
  }
}
