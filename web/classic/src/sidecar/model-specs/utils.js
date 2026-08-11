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
// label 作为 i18n key，渲染时通过 t(label) 取当前语言文案
export const CAPABILITY_OPTIONS = [
  { value: 'vision', label: 'Vision' },
  { value: 'reasoning', label: 'Reasoning' },
  { value: 'code', label: 'Code' },
  { value: 'tool-use', label: 'Tool Use' },
  { value: 'audio', label: 'Audio' },
  { value: 'image', label: 'Image' },
  { value: 'video', label: 'Video' },
  { value: 'search', label: 'Search' },
  { value: 'long-context', label: 'Long Context' },
];

const CAPABILITY_LABEL_MAP = Object.fromEntries(
  CAPABILITY_OPTIONS.map((option) => [option.value, option.label])
);

/**
 * 将能力值（如 'vision'）转换为 i18n key（如 'Vision'）。
 * 若未命中已知能力，则原样返回，便于作为兜底显示。
 */
export function getCapabilityLabel(value) {
  return CAPABILITY_LABEL_MAP[value] ?? value;
}

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
