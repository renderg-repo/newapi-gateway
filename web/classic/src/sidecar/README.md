# Sidecar 模式开发说明

## 什么是 Sidecar 模式？

Sidecar 模式是一种最小侵入式的功能扩展方式，新功能完全隔离在 `sidecar/` 目录下，只在必要的地方进行最小修改来集成功能。

## 为什么使用 Sidecar 模式？

1. ✅ **完全解耦** - 新功能代码与主项目完全隔离
2. ✅ **易于维护** - Sidecar 模块可以独立开发、测试、部署
3. ✅ **低冲突风险** - 最小化对主代码的修改，减少上游合并冲突
4. ✅ **可插拔** - 可以轻松启用/禁用 Sidecar 功能

## 目录结构

```
web/classic/src/sidecar/
├── index.js              # Sidecar 统一导出
├── README.md             # 本文档
└── model-specs/          # 模型规格 Sidecar 模块
    ├── register.js       # Sidecar 注册器
    ├── index.jsx         # 主组件
    ├── utils.js          # 工具函数（独立）
    └── components/
        ├── ModelSpecsColumnDefs.jsx
        ├── ModelSpecsTable.jsx
        └── modals/
            └── ModelSpecEditModal.jsx
```

## Sidecar 开发规范

### 1. 目录隔离
- ✅ 所有 Sidecar 代码放在 `sidecar/` 目录下
- ✅ 按功能模块划分子目录（如 `model-specs/`）

### 2. 依赖最小化
- ✅ Sidecar 内部可以使用主项目的公共组件和 helpers
- ✅ 但 Sidecar 之间应该互相独立
- ✅ Sidecar 特有的工具函数放在自己目录内

### 3. 注册模式
每个 Sidecar 模块应该提供一个 `register.js`：

```javascript
// 导出组件
export { MyComponent };

// 导出配置
export const myTab = {
  key: 'my-feature',
  label: '我的功能',
  component: MyComponent,
};

// 导出获取函数
export function getMyTab() {
  return myTab;
}
```

### 4. 集成方式（最小侵入）

#### 方式 A：动态 require（推荐）
```javascript
let MySidecar = null;
try {
  const sidecar = require('../../sidecar/my-feature/register');
  MySidecar = sidecar.MyComponent;
} catch (e) {
  console.log('[Sidecar] 功能未启用');
}

// 使用时
{MySidecar && <MySidecar />}
```

#### 方式 B：配置化集成
```javascript
// 在主项目的一个集中配置文件中
import { modelSpecs } from '../sidecar';

const enabledSidecars = [
  modelSpecs,
];
```

### 5. 侵入点限制
只在以下地方允许最小化修改：
- ✅ 路由/页面组件中动态导入 Sidecar
- ✅ 导航菜单中条件性添加 Sidecar 入口
- ❌ 禁止修改主项目的核心逻辑文件
- ❌ 禁止在主项目的 utils/constants 中添加 Sidecar 特有代码

## 模型规格 Sidecar 示例

### 目录结构
```
sidecar/model-specs/
├── register.js          # 注册器
├── index.jsx            # 主组件
├── utils.js             # 独立工具函数
└── components/
    ├── ...
```

### 集成方式
在 `pages/Model/index.jsx` 中：
```javascript
// 最小侵入：动态导入
let ModelSpecs = null;
let modelSpecsTabConfig = null;
try {
  const sidecar = require('../../sidecar/model-specs/register');
  ModelSpecs = sidecar.ModelSpecs;
  modelSpecsTabConfig = sidecar.getModelSpecsTab();
} catch (e) {
  console.log('[Sidecar] 模型规格功能未启用');
}

// 条件渲染
if (ModelSpecs && modelSpecsTabConfig) {
  // 添加标签页
}
```

## 新增 Sidecar 模块步骤

1. 在 `sidecar/` 下创建新目录
2. 创建 `register.js` 注册器
3. 实现功能代码（完全在目录内）
4. 在需要集成的地方动态导入
5. 添加条件性渲染逻辑

## 禁用 Sidecar

只需删除或重命名对应目录即可，主项目不会受影响。
