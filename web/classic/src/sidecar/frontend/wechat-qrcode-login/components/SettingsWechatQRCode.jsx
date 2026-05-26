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

export default function SettingsWechatQRCode(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    WeChatAppID: '',
    WeChatAppSecret: '',
    WeChatReceiveToken: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        WeChatAppID: props.options.WeChatAppID || '',
        WeChatAppSecret: props.options.WeChatAppSecret || '',
        WeChatReceiveToken: props.options.WeChatReceiveToken || '',
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitWechatQRCode = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      const options = [];

      if (inputs.WeChatAppID !== '') {
        options.push({ key: 'WeChatAppID', value: inputs.WeChatAppID });
      }
      if (
        inputs.WeChatAppSecret &&
        inputs.WeChatAppSecret !== '' &&
        inputs.WeChatAppSecret !== originInputs.WeChatAppSecret
      ) {
        options.push({ key: 'WeChatAppSecret', value: inputs.WeChatAppSecret });
      }
      if (inputs.WeChatReceiveToken !== '') {
        options.push({ key: 'WeChatReceiveToken', value: inputs.WeChatReceiveToken });
      }

      if (options.length === 0) {
        showSuccess(t('没有需要更新的设置'));
        setLoading(false);
        return;
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
        <Form.Section text={t('微信扫码登录设置')}>
          <Text>
            {t('使用微信公众号官方接口，用户扫码完成登录')}
          </Text>
          <Banner
            type='info'
            description={`${t('回调地址')}：${props.options.ServerAddress ? removeTrailingSlash(props.options.ServerAddress) : t('网站地址')}/api/weixin/receiveMessage`}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }} style={{ marginTop: 16 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WeChatAppID'
                label={t('微信公众号 AppID')}
                placeholder={t('微信公众号的 AppID')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WeChatAppSecret'
                label={t('微信公众号 AppSecret')}
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WeChatReceiveToken'
                label={t('消息回调 Token')}
                placeholder={t('微信公众号消息回调验证 Token')}
              />
            </Col>
          </Row>
          <Button onClick={submitWechatQRCode} style={{ marginTop: 16 }}>
            {t('保存微信扫码登录设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
