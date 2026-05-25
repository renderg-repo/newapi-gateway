import React from 'react';
import { Modal } from '@douyinfe/semi-ui';
import { API, showError } from '../../../helpers';

export async function wechatPayTopUp(params) {
  const {
    topUpCount,
    minTopUp,
    setPaymentLoading,
    t,
    getUserQuota,
    setOpenHistory,
  } = params;

  if (topUpCount < minTopUp) {
    showError(t('充值数量不能小于') + minTopUp);
    return;
  }

  setPaymentLoading(true);
  try {
    const res = await API.post('/api/user/wechatpay/pay', {
      amount: parseInt(topUpCount),
    });
    if (res !== undefined) {
      const { message, data } = res.data;
      if (message === 'success' && data?.code_url) {
        const qrModal = Modal.info({
          title: (
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <span>{t('微信支付')}</span>
            </div>
          ),
          content: (
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                padding: '8px 0',
              }}
            >
              <p
                style={{
                  margin: '0 0 16px 0',
                  fontSize: 14,
                  color: 'var(--semi-color-text-1)',
                }}
              >
                {t('请使用微信扫描二维码完成支付')}
              </p>
              <div
                style={{
                  padding: 16,
                  background: 'var(--semi-color-fill-0)',
                  borderRadius: 12,
                  marginBottom: 16,
                }}
              >
                <img
                  src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(data.code_url)}`}
                  alt='wechat pay qr'
                  style={{
                    width: 200,
                    height: 200,
                    display: 'block',
                    borderRadius: 8,
                  }}
                />
              </div>
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 4,
                }}
              >
                <span
                  style={{
                    fontSize: 12,
                    color: 'var(--semi-color-text-2)',
                  }}
                >
                  {t('订单号')}
                </span>
                <span
                  style={{
                    fontSize: 13,
                    color: 'var(--semi-color-text-1)',
                    fontFamily: 'monospace',
                    letterSpacing: 0.5,
                  }}
                >
                  {data.trade_no}
                </span>
              </div>
            </div>
          ),
          okText: t('已完成支付'),
          cancelText: t('关闭'),
          onOk: () => {
            getUserQuota();
            setOpenHistory(true);
          },
          onCancel: () => {
            qrModal.destroy();
          },
        });
      } else {
        showError(data || t('支付请求失败'));
      }
    } else {
      showError(res);
    }
  } catch (e) {
    showError(t('支付请求失败'));
  } finally {
    setPaymentLoading(false);
  }
}

export async function alipayTopUp(params) {
  const { topUpCount, minTopUp, setPaymentLoading, t } = params;

  if (topUpCount < minTopUp) {
    showError(t('充值数量不能小于') + minTopUp);
    return;
  }

  setPaymentLoading(true);
  try {
    const res = await API.post('/api/user/alipay/pay', {
      amount: parseInt(topUpCount),
    });
    if (res !== undefined) {
      const { message, data } = res.data;
      if (message === 'success' && data?.pay_url) {
        window.open(data.pay_url, '_blank');
      } else {
        showError(data || t('支付请求失败'));
      }
    } else {
      showError(res);
    }
  } catch (e) {
    showError(t('支付请求失败'));
  } finally {
    setPaymentLoading(false);
  }
}
