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

import React, { useContext, useEffect, useState, useMemo } from 'react';
import {
  Button,
  Typography,
} from '@douyinfe/semi-ui';
import { API, showError, copy, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { API_ENDPOINTS } from '../../constants/common.constant';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import {
  IconGithubLogo,
  IconPlay,
  IconFile,
  IconCopy,
  IconBolt,
  IconShield,
  IconHistogram,
} from '@douyinfe/semi-icons';
import { Link } from 'react-router-dom';
import NoticeModal from '../../components/layout/NoticeModal';
import {
  Moonshot,
  OpenAI,
  XAI,
  Zhipu,
  Volcengine,
  Cohere,
  Claude,
  Gemini,
  Suno,
  Minimax,
  Wenxin,
  Spark,
  Qingyan,
  DeepSeek,
  Qwen,
  Midjourney,
  Grok,
  AzureAI,
  Hunyuan,
  Xinference,
} from '@lobehub/icons';

const { Text } = Typography;

/* ================================================================
   Particle system — sparse, subtle twinkling dots
   ================================================================ */
const PARTICLE_COUNT = 24;
const particles = Array.from({ length: PARTICLE_COUNT }, (_, i) => ({
  id: i,
  left: `${Math.random() * 100}%`,
  top: `${Math.random() * 100}%`,
  size: Math.random() * 2 + 1,
  delay: Math.random() * 6,
  duration: Math.random() * 4 + 4,
  opacity: Math.random() * 0.4 + 0.15,
}));

/* ================================================================
   Provider logo list (re-used for seamless marquee)
   ================================================================ */
const PROVIDERS = [
  { Component: Moonshot, color: false },
  { Component: OpenAI, color: false },
  { Component: XAI, color: false },
  { Component: Zhipu, color: true },
  { Component: Volcengine, color: true },
  { Component: Cohere, color: true },
  { Component: Claude, color: true },
  { Component: Gemini, color: true },
  { Component: Suno, color: false },
  { Component: Minimax, color: true },
  { Component: Wenxin, color: true },
  { Component: Spark, color: true },
  { Component: Qingyan, color: true },
  { Component: DeepSeek, color: true },
  { Component: Qwen, color: true },
  { Component: Midjourney, color: false },
  { Component: Grok, color: false },
  { Component: AzureAI, color: true },
  { Component: Hunyuan, color: true },
  { Component: Xinference, color: true },
];

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;
  const isDark = actualTheme === 'dark';
  const docsLink = statusState?.status?.docs_link || '';
  const serverAddress =
    statusState?.status?.server_address || `${window.location.origin}`;
  const endpointItems = API_ENDPOINTS.map((e) => ({ value: e }));
  const [endpointIndex, setEndpointIndex] = useState(0);
  const isChinese = i18n.language.startsWith('zh');

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);

      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  const handleCopyBaseURL = async () => {
    const ok = await copy(serverAddress);
    if (ok) {
      showSuccess(t('已复制到剪切板'));
    }
  };

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };

    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      setEndpointIndex((prev) => (prev + 1) % endpointItems.length);
    }, 3000);
    return () => clearInterval(timer);
  }, [endpointItems.length]);

  /* Current endpoint label */
  const currentEndpoint = endpointItems[endpointIndex]?.value || '/v1/chat/completions';

  /* Provider logo size */
  const logoSize = isMobile ? 22 : 30;
  const logoColor = isDark ? 'rgba(255, 255, 255, 0.55)' : 'rgba(0, 0, 0, 0.45)';

  /* Marquee content — duplicated for seamless infinite scroll */
  const marqueeLogos = useMemo(() => (
    <>
      {PROVIDERS.map(({ Component, color }, i) => (
        <div
          key={`a-${i}`}
          className='flex-shrink-0 flex items-center justify-center'
          style={{ width: logoSize + 20 }}
        >
          {color ? (
            <Component size={logoSize} />
          ) : (
            <Component size={logoSize} color={logoColor} />
          )}
        </div>
      ))}
      <div
        className='flex-shrink-0 flex items-center justify-center'
        style={{ width: logoSize + 20 }}
      >
        <Typography.Text
          className={`!text-base sm:!text-lg md:!text-xl font-bold ${isDark ? 'text-white/55' : 'text-black/45'}`}
        >
          30+
        </Typography.Text>
      </div>
    </>
  ), [logoSize, logoColor, isDark]);

  const currentYear = new Date().getFullYear();

  return (
    <div className='w-full h-full overflow-hidden'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='w-full h-full overflow-hidden'>
          {/* ==================== Hero Section ==================== */}
          <div
            className={`w-full h-[calc(100dvh-64px)] relative overflow-hidden flex flex-col ${isDark ? 'home-hero-dark' : 'home-hero-light'}`}
          >
            {/* ---- Mesh Gradient Background ---- */}
            <div className='mesh-gradient-bg' aria-hidden='true' />

            {/* ---- Subtle particles ---- */}
            <div className='particles-layer' aria-hidden='true'>
              {particles.map((p) => (
                <span
                  key={p.id}
                  className='particle-dot'
                  style={{
                    left: p.left,
                    top: p.top,
                    width: p.size,
                    height: p.size,
                    animationDelay: `${p.delay}s`,
                    animationDuration: `${p.duration}s`,
                    opacity: p.opacity,
                  }}
                />
              ))}
            </div>

            {/* ---- Center Content ---- */}
            <div className='relative z-10 flex-1 flex items-center justify-center px-4 pt-16'>
              <div className='flex flex-col items-center justify-center text-center max-w-4xl mx-auto w-full'>
                {/* Brand name */}
                <div className='hero-glow-text text-3xl md:text-4xl lg:text-5xl font-bold mb-2 md:mb-3 tracking-widest'>
                  New API
                </div>

                {/* Main title */}
                <h1
                  className={`text-2xl md:text-3xl lg:text-4xl font-bold leading-tight ${isChinese ? 'tracking-wide md:tracking-wider' : ''}`}
                  style={{ color: isDark ? '#ffffff' : '#111827' }}
                >
                  <>{t('统一的')}</>
                  <br />
                  <>{t('大模型接口网关')}</>
                </h1>

                {/* Subtitle */}
                <p
                  className={`text-sm md:text-base mt-2 md:mt-3 max-w-lg ${isDark ? 'text-white/55' : 'text-black/55'}`}
                >
                  {t('更好的价格，更好的稳定性，只需要将模型基址替换为：')}
                </p>

                {/* Glassmorphism Code Block */}
                <div className='mt-4 md:mt-5 w-full max-w-lg mx-auto px-2'>
                  <div
                    className={`code-block-glass flex items-center gap-3 px-4 py-2.5 md:px-5 md:py-3 rounded-xl border ${isDark ? 'border-white/[0.08]' : 'border-black/[0.06]'}`}
                  >
                    {/* URL text */}
                    <span
                      className={`text-sm md:text-base font-mono truncate shrink-0 ${isDark ? 'text-white/80' : 'text-gray-700'}`}
                    >
                      {serverAddress}
                    </span>

                    {/* Endpoint tag */}
                    <span
                      className={`text-xs md:text-sm font-mono px-2 py-0.5 rounded-md shrink-0 ${isDark ? 'bg-white/[0.08] text-white/50' : 'bg-black/[0.05] text-gray-400'}`}
                    >
                      {currentEndpoint}
                    </span>

                    {/* Spacer */}
                    <span className='flex-1' />

                    {/* Copy button */}
                    <button
                      onClick={handleCopyBaseURL}
                      className={`shrink-0 w-8 h-8 flex items-center justify-center rounded-lg transition-all duration-300 group ${isDark ? 'hover:bg-white/10' : 'hover:bg-black/5'}`}
                      title={t('复制')}
                    >
                      <IconCopy
                        size='small'
                        className={`transition-all duration-300 ${isDark ? 'text-white/50 group-hover:text-white group-hover:drop-shadow-[0_0_6px_rgba(38,150,255,0.6)]' : 'text-gray-400 group-hover:text-gray-700'}`}
                      />
                    </button>
                  </div>
                </div>

                {/* Action buttons */}
                <div className='flex flex-row gap-4 justify-center items-center mt-5'>
                  <Link to='/console'>
                    <Button
                      size={isMobile ? 'default' : 'large'}
                      className={`!rounded-full px-8 py-2 ${isDark ? 'hero-primary-btn hero-btn-glow' : '!bg-[#2156b3] hover:!bg-[#3070e1]'}`}
                      style={isDark ? undefined : { color: '#ffffff' }}
                      icon={<IconPlay />}
                    >
                      {t('获取密钥')}
                    </Button>
                  </Link>
                  {isDemoSiteMode && statusState?.status?.version ? (
                    <Button
                      size={isMobile ? 'default' : 'large'}
                      className={`flex items-center !rounded-full px-6 py-2 !bg-transparent !border ${isDark ? '!border-white/20 !text-white hover:!bg-white/10' : '!border-gray-300 !text-gray-700 hover:!bg-gray-100'}`}
                      icon={<IconGithubLogo />}
                      onClick={() =>
                        window.open(
                          'https://github.com/QuantumNous/new-api',
                          '_blank',
                        )
                      }
                    >
                      {statusState.status.version}
                    </Button>
                  ) : (
                    docsLink && (
                      <Button
                        size={isMobile ? 'default' : 'large'}
                        className={`flex items-center !rounded-full px-6 py-2 !bg-transparent !border ${isDark ? '!border-[rgba(38,150,255,0.35)] !text-white hover:!bg-[rgba(38,150,255,0.08)]' : '!border-gray-300 !text-gray-700 hover:!bg-gray-100'}`}
                        icon={<IconFile />}
                        onClick={() => window.open(docsLink, '_blank')}
                      >
                        {t('文档')}
                      </Button>
                    )
                  )}
                </div>

                {/* ==================== Value Cards ==================== */}
                <div className='mt-6 md:mt-8 w-full'>
                  <div className='flex flex-row items-stretch justify-center gap-2 md:gap-3 max-w-3xl mx-auto px-2'>
                    {/* Card 1 — Speed */}
                    <div
                      className={`flex-1 flex items-center gap-2.5 px-3 py-2.5 rounded-lg transition-all duration-300 ${isDark ? 'bg-white/[0.03] border border-white/[0.06] hover:bg-white/[0.05] hover:border-white/[0.10]' : 'bg-white/60 border border-black/[0.04] hover:bg-white/80 hover:border-black/[0.08]'}`}
                      style={{ backdropFilter: 'blur(8px)' }}
                    >
                      <div
                        className={`w-8 h-8 flex items-center justify-center rounded-lg shrink-0 ${isDark ? 'bg-[rgba(38,150,255,0.12)] text-[#2696ff]' : 'bg-blue-50 text-[#2156b3]'}`}
                      >
                        <IconBolt size='small' />
                      </div>
                      <div className='text-left min-w-0'>
                        <div
                          className={`text-xs md:text-sm font-semibold ${isDark ? 'text-white/85' : 'text-gray-800'}`}
                        >
                          {t('极速响应')}
                        </div>
                        <div
                          className={`text-[11px] md:text-xs mt-0.5 ${isDark ? 'text-white/40' : 'text-gray-500'}`}
                        >
                          {t('毫秒级低延迟中转网关')}
                        </div>
                      </div>
                    </div>

                    {/* Card 2 — Reliability */}
                    <div
                      className={`flex-1 flex items-center gap-2.5 px-3 py-2.5 rounded-lg transition-all duration-300 ${isDark ? 'bg-white/[0.03] border border-white/[0.06] hover:bg-white/[0.05] hover:border-white/[0.10]' : 'bg-white/60 border border-black/[0.04] hover:bg-white/80 hover:border-black/[0.08]'}`}
                      style={{ backdropFilter: 'blur(8px)' }}
                    >
                      <div
                        className={`w-8 h-8 flex items-center justify-center rounded-lg shrink-0 ${isDark ? 'bg-[rgba(38,150,255,0.12)] text-[#2696ff]' : 'bg-blue-50 text-[#2156b3]'}`}
                      >
                        <IconShield size='small' />
                      </div>
                      <div className='text-left min-w-0'>
                        <div
                          className={`text-xs md:text-sm font-semibold ${isDark ? 'text-white/85' : 'text-gray-800'}`}
                        >
                          {t('高可用性')}
                        </div>
                        <div
                          className={`text-[11px] md:text-xs mt-0.5 ${isDark ? 'text-white/40' : 'text-gray-500'}`}
                        >
                          {t('99.9% 系统正常运行时间保证')}
                        </div>
                      </div>
                    </div>

                    {/* Card 3 — Cost */}
                    <div
                      className={`flex-1 flex items-center gap-2.5 px-3 py-2.5 rounded-lg transition-all duration-300 ${isDark ? 'bg-white/[0.03] border border-white/[0.06] hover:bg-white/[0.05] hover:border-white/[0.10]' : 'bg-white/60 border border-black/[0.04] hover:bg-white/80 hover:border-black/[0.08]'}`}
                      style={{ backdropFilter: 'blur(8px)' }}
                    >
                      <div
                        className={`w-8 h-8 flex items-center justify-center rounded-lg shrink-0 ${isDark ? 'bg-[rgba(38,150,255,0.12)] text-[#2696ff]' : 'bg-blue-50 text-[#2156b3]'}`}
                      >
                        <IconHistogram size='small' />
                      </div>
                      <div className='text-left min-w-0'>
                        <div
                          className={`text-xs md:text-sm font-semibold ${isDark ? 'text-white/85' : 'text-gray-800'}`}
                        >
                          {t('成本透明')}
                        </div>
                        <div
                          className={`text-[11px] md:text-xs mt-0.5 ${isDark ? 'text-white/40' : 'text-gray-500'}`}
                        >
                          {t('统一多模型计费与实时账单')}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* ==================== Provider Logos ==================== */}
                <div className='mt-6 md:mt-8 w-full'>
                  <div className='flex items-center mb-3 md:mb-4 justify-center'>
                    <Text
                      className={`text-base md:text-lg lg:text-xl font-bold ${isDark ? 'text-white/55' : 'text-black/45'}`}
                    >
                      {t('支持众多的大模型供应商')}
                    </Text>
                  </div>

                  {/* Infinite-scroll logo marquee */}
                  <div className='relative w-full overflow-hidden py-1'>
                    {/* Left fade mask */}
                    <div
                      className={`absolute left-0 top-0 bottom-0 w-10 sm:w-14 md:w-20 z-10 pointer-events-none ${isDark ? 'home-fade-left-dark' : 'home-fade-left-light'}`}
                    />
                    {/* Right fade mask */}
                    <div
                      className={`absolute right-0 top-0 bottom-0 w-10 sm:w-14 md:w-20 z-10 pointer-events-none ${isDark ? 'home-fade-right-dark' : 'home-fade-right-light'}`}
                    />

                    <div className='logo-marquee-track flex items-center'>
                      {marqueeLogos}
                      {marqueeLogos}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* ---- Inline Footer (no scroll) ---- */}
            <div className='relative z-10 flex-shrink-0 py-3 px-4 text-center'>
              <Typography.Text
                className={`text-xs ${isDark ? 'text-white/30' : 'text-black/30'}`}
              >
                北京润德骥图科技有限公司 Copyright ©{currentYear} spacehpc.com 备案号：
                <a
                  href='https://beian.miit.gov.cn'
                  target='_blank'
                  rel='noopener noreferrer'
                  className={`${isDark ? 'text-white/40 hover:text-white/60' : 'text-black/40 hover:text-black/60'} transition-colors`}
                >
                  京ICP备2020049003号-1
                </a>
              </Typography.Text>
            </div>
          </div>
        </div>
      ) : (
        <div className='w-full h-full overflow-hidden'>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              className='w-full h-screen border-none'
            />
          ) : (
            <div
              className='mt-[60px]'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default Home;
