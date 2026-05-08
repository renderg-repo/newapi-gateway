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

import React, { useEffect, useState, useRef } from 'react';
import {
  Banner,
  Button,
  Form,
  Row,
  Col,
  Typography,
  Spin,
} from '@douyinfe/semi-ui';
const { Text } = Typography;
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
} from '../../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsWechatPay(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    WechatPayEnabled: false,
    WechatPayAppId: '',
    WechatPayMchId: '',
    WechatPayApiKey: '',
    WechatPayUnitPrice: 7.3,
    WechatPayMinTopUp: 1,
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        WechatPayEnabled:
          props.options.WechatPayEnabled === 'true' ||
          props.options.WechatPayEnabled === true,
        WechatPayAppId: props.options.WechatPayAppId || '',
        WechatPayMchId: props.options.WechatPayMchId || '',
        WechatPayApiKey: props.options.WechatPayApiKey || '',
        WechatPayUnitPrice:
          props.options.WechatPayUnitPrice !== undefined
            ? parseFloat(props.options.WechatPayUnitPrice)
            : 7.3,
        WechatPayMinTopUp:
          props.options.WechatPayMinTopUp !== undefined
            ? parseFloat(props.options.WechatPayMinTopUp)
            : 1,
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitWechatPay = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      const options = [];

      if (
        originInputs['WechatPayEnabled'] !== inputs.WechatPayEnabled &&
        inputs.WechatPayEnabled !== undefined
      ) {
        options.push({
          key: 'WechatPayEnabled',
          value: inputs.WechatPayEnabled ? 'true' : 'false',
        });
      }
      if (inputs.WechatPayAppId !== '') {
        options.push({ key: 'WechatPayAppId', value: inputs.WechatPayAppId });
      }
      if (inputs.WechatPayMchId !== '') {
        options.push({ key: 'WechatPayMchId', value: inputs.WechatPayMchId });
      }
      if (
        inputs.WechatPayApiKey &&
        inputs.WechatPayApiKey !== '' &&
        inputs.WechatPayApiKey !== originInputs.WechatPayApiKey
      ) {
        options.push({ key: 'WechatPayApiKey', value: inputs.WechatPayApiKey });
      }
      if (
        inputs.WechatPayUnitPrice !== undefined &&
        inputs.WechatPayUnitPrice !== null
      ) {
        options.push({
          key: 'WechatPayUnitPrice',
          value: inputs.WechatPayUnitPrice.toString(),
        });
      }
      if (
        inputs.WechatPayMinTopUp !== undefined &&
        inputs.WechatPayMinTopUp !== null
      ) {
        options.push({
          key: 'WechatPayMinTopUp',
          value: inputs.WechatPayMinTopUp.toString(),
        });
      }

      const requestQueue = options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      );

      const results = await Promise.all(requestQueue);

      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        setOriginInputs({ ...inputs });
        props.refresh?.();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={t('微信支付设置')}>
          <Text>
            {t('使用微信支付官方 Native 支付接口，用户扫码完成充值')}
          </Text>
          <Banner
            type='info'
            description={`${t('回调地址')}：${props.options.ServerAddress ? removeTrailingSlash(props.options.ServerAddress) : t('网站地址')}/api/wechatpay/notify`}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='WechatPayEnabled'
                size='default'
                checkedText='｜'
                uncheckedText='〇'
                label={t('启用微信支付')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayAppId'
                label={t('微信 AppID')}
                placeholder={t('微信公众号或小程序的 AppID')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayMchId'
                label={t('微信商户号')}
                placeholder={t('微信支付商户号')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayApiKey'
                label={t('API 密钥')}
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='WechatPayUnitPrice'
                precision={2}
                label={t('充值价格（元/单位）')}
                placeholder={t('例如：7.3')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='WechatPayMinTopUp'
                label={t('最低充值金额')}
                placeholder={t('例如：1')}
              />
            </Col>
          </Row>
          <Button onClick={submitWechatPay} style={{ marginTop: 16 }}>
            {t('更新微信支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
