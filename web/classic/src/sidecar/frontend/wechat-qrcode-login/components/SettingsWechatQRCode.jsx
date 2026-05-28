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
    WeChatHpcServerAddress: '',
    WeChatHpcCallbackToken: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        WeChatHpcServerAddress: props.options.WeChatHpcServerAddress || '',
        WeChatHpcCallbackToken: props.options.WeChatHpcCallbackToken || '',
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
    setLoading(true);
    try {
      const options = [];

      if (inputs.WeChatHpcServerAddress !== '') {
        options.push({
          key: 'WeChatHpcServerAddress',
          value: removeTrailingSlash(inputs.WeChatHpcServerAddress)
        });
      }

      if (inputs.WeChatHpcCallbackToken !== '' &&
          inputs.WeChatHpcCallbackToken !== originInputs.WeChatHpcCallbackToken) {
        options.push({
          key: 'WeChatHpcCallbackToken',
          value: inputs.WeChatHpcCallbackToken
        });
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
            {t('通过智算系统的微信公众号接口，用户扫码完成登录')}
          </Text>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }} style={{ marginTop: 16 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WeChatHpcServerAddress'
                label={t('智算系统服务器地址')}
                placeholder='例如：https://hpc.example.com'
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WeChatHpcCallbackToken'
                label={t('回调接口验证Token')}
                placeholder='输入一个安全的Token'
                type='password'
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
