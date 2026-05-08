import { pluginRegistry } from '../../../helpers/pluginRegistry';
import SettingsWechatPay from './components/SettingsWechatPay';
import SettingsAlipay from './components/SettingsAlipay';

// 注册支付网关元数据
pluginRegistry.registerPaymentGateway('wechatpay_native', {
  icon: 'SiWechat',
  iconColor: '#07C160',
  category: 'native',
});

pluginRegistry.registerPaymentGateway('alipay_page', {
  icon: 'SiAlipay',
  iconColor: '#1677FF',
  category: 'native',
});

// 注册设置页面 Card
pluginRegistry.registerSettingCard(SettingsWechatPay);
pluginRegistry.registerSettingCard(SettingsAlipay);
