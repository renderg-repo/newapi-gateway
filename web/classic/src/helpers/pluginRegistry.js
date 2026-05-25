import React from 'react';

class PluginRegistry {
  constructor() {
    // 支付网关元数据: type -> { icon, iconColor, category, ... }
    this.paymentGateways = new Map();
    // 自定义充值动作: type -> function
    this.topupActions = new Map();
    // 设置页面组件列表
    this.settingCards = [];
  }

  /**
   * 注册支付网关元数据
   * @param {string} type - 支付类型标识，如 'wechatpay_native'
   * @param {object} config
   * @param {string} config.icon - 图标组件名，如 'SiWechat'
   * @param {string} config.iconColor - 图标颜色，如 '#07C160'
   * @param {string} config.category - 分类
   */
  registerPaymentGateway(type, config) {
    this.paymentGateways.set(type, {
      icon: 'CreditCard',
      iconColor: null,
      category: 'other',
      ...config,
    });
  }

  /**
   * 注册充值动作（点击支付按钮时执行，替代 preTopUp 中的硬编码分支）
   * @param {string} type - 支付类型标识
   * @param {function} actionFn - 动作函数
   */
  registerTopupAction(type, actionFn) {
    this.topupActions.set(type, actionFn);
  }

  /**
   * 注册设置页面 Card 组件
   * @param {React.ComponentType} component
   */
  registerSettingCard(component) {
    this.settingCards.push(component);
  }

  getPaymentGateway(type) {
    return this.paymentGateways.get(type);
  }

  hasTopupAction(type) {
    return this.topupActions.has(type);
  }

  getTopupAction(type) {
    return this.topupActions.get(type);
  }

  getAllSettingCards() {
    return this.settingCards;
  }
}

export const pluginRegistry = new PluginRegistry();
