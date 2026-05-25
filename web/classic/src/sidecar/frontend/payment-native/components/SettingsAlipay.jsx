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

export default function SettingsAlipay(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    AlipayEnabled: false,
    AlipayAppId: '',
    AlipayPrivateKey: '',
    AlipayPublicKey: '',
    AlipayGatewayUrl: '',
    AlipayUnitPrice: 7.3,
    AlipayMinTopUp: 1,
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        AlipayEnabled:
          props.options.AlipayEnabled === 'true' ||
          props.options.AlipayEnabled === true,
        AlipayAppId: props.options.AlipayAppId || '',
        AlipayPrivateKey: props.options.AlipayPrivateKey || '',
        AlipayPublicKey: props.options.AlipayPublicKey || '',
        AlipayGatewayUrl: props.options.AlipayGatewayUrl || '',
        AlipayUnitPrice:
          props.options.AlipayUnitPrice !== undefined
            ? parseFloat(props.options.AlipayUnitPrice)
            : 7.3,
        AlipayMinTopUp:
          props.options.AlipayMinTopUp !== undefined
            ? parseFloat(props.options.AlipayMinTopUp)
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

  const submitAlipay = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      const options = [];

      if (
        originInputs['AlipayEnabled'] !== inputs.AlipayEnabled &&
        inputs.AlipayEnabled !== undefined
      ) {
        options.push({
          key: 'AlipayEnabled',
          value: inputs.AlipayEnabled ? 'true' : 'false',
        });
      }
      if (inputs.AlipayAppId !== '') {
        options.push({ key: 'AlipayAppId', value: inputs.AlipayAppId });
      }
      if (
        inputs.AlipayPrivateKey &&
        inputs.AlipayPrivateKey !== '' &&
        inputs.AlipayPrivateKey !== originInputs.AlipayPrivateKey
      ) {
        options.push({ key: 'AlipayPrivateKey', value: inputs.AlipayPrivateKey });
      }
      if (
        inputs.AlipayPublicKey &&
        inputs.AlipayPublicKey !== '' &&
        inputs.AlipayPublicKey !== originInputs.AlipayPublicKey
      ) {
        options.push({ key: 'AlipayPublicKey', value: inputs.AlipayPublicKey });
      }
      if (inputs.AlipayGatewayUrl !== '') {
        options.push({ key: 'AlipayGatewayUrl', value: inputs.AlipayGatewayUrl });
      }
      if (
        inputs.AlipayUnitPrice !== undefined &&
        inputs.AlipayUnitPrice !== null
      ) {
        options.push({
          key: 'AlipayUnitPrice',
          value: inputs.AlipayUnitPrice.toString(),
        });
      }
      if (
        inputs.AlipayMinTopUp !== undefined &&
        inputs.AlipayMinTopUp !== null
      ) {
        options.push({
          key: 'AlipayMinTopUp',
          value: inputs.AlipayMinTopUp.toString(),
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
        <Form.Section text={t('支付宝设置')}>
          <Text>
            {t('使用支付宝官方电脑网站支付接口，用户跳转支付宝收银台完成充值')}
          </Text>
          <Banner
            type='info'
            description={`${t('回调地址')}：${props.options.ServerAddress ? removeTrailingSlash(props.options.ServerAddress) : t('网站地址')}/api/alipay/notify`}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='AlipayEnabled'
                size='default'
                checkedText='｜'
                uncheckedText='〇'
                label={t('启用支付宝')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='AlipayAppId'
                label={t('支付宝 AppID')}
                placeholder={t('支付宝开放平台的应用 ID')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='AlipayGatewayUrl'
                label={t('网关地址')}
                placeholder={t('https://openapi.alipay.com/gateway.do')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='AlipayPrivateKey'
                label={t('应用私钥')}
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='AlipayPublicKey'
                label={t('支付宝公钥')}
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='AlipayUnitPrice'
                precision={2}
                label={t('充值价格（元/单位）')}
                placeholder={t('例如：7.3')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='AlipayMinTopUp'
                label={t('最低充值金额')}
                placeholder={t('例如：1')}
              />
            </Col>
          </Row>
          <Button onClick={submitAlipay} style={{ marginTop: 16 }}>
            {t('更新支付宝设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
