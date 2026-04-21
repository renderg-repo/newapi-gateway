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
import { Typography } from '@douyinfe/semi-ui';
import { API, showError, copy, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { API_ENDPOINTS } from '../../constants/common.constant';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import NoticeModal from '../../components/layout/NoticeModal';
import {
  Zap,
  Shield,
  BarChart3,
  Copy,
  Play,
  FileText,
  Github,
} from 'lucide-react';
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
  { Component: Moonshot, colorDark: '#8B9EFF', colorLight: '#16191E' },
  { Component: OpenAI, colorDark: '#fff', colorLight: '#000' },
  { Component: XAI, colorDark: '#fff', colorLight: '#333' },
  { Component: Zhipu, colorDark: '#5c7cff', colorLight: '#3859FF' },
  { Component: Volcengine, colorDark: '#fff', colorLight: '#333' },
  { Component: Cohere, colorDark: '#6abf9e', colorLight: '#39594D' },
  { Component: Claude, colorDark: '#e89a7d', colorLight: '#D97757' },
  { Component: Gemini, colorDark: '#fff', colorLight: '#333' },
  { Component: Suno, colorDark: '#FF6B6B', colorLight: '#000' },
  { Component: Minimax, colorDark: '#f56f82', colorLight: '#F23F5D' },
  { Component: Wenxin, colorDark: '#4a9eff', colorLight: '#167ADF' },
  { Component: Spark, colorDark: '#4a9fff', colorLight: '#0070f0' },
  { Component: Qingyan, colorDark: '#4d7aff', colorLight: '#1041F3' },
  { Component: DeepSeek, colorDark: '#7a8fff', colorLight: '#4D6BFE' },
  { Component: Qwen, colorDark: '#8a7fff', colorLight: '#615ced' },
  { Component: Midjourney, colorDark: '#fff', colorLight: '#333' },
  { Component: Grok, colorDark: '#fff', colorLight: '#000' },
  { Component: AzureAI, colorDark: '#4da6ff', colorLight: '#000' },
  { Component: Hunyuan, colorDark: '#4d8fff', colorLight: '#0053e0' },
  { Component: Xinference, colorDark: '#a66aff', colorLight: '#781ff5' },
];

/* ================================================================
   Font stack — SF Pro / Inter / system fallback
   ================================================================ */
const fontStack = {
  fontFamily:
    '-apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "Inter", "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif',
};

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
  // i18n language info available if needed

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

  /* Marquee content — duplicated for seamless infinite scroll */
  const marqueeLogos = useMemo(
    () => (
      <>
        {PROVIDERS.map(({ Component, colorDark, colorLight }, i) => (
          <div
            key={`a-${i}`}
            className='logo-item-wrapper flex-shrink-0 flex items-center justify-center'
            style={{
              width: logoSize + 40,
              '--logo-color-dark': colorDark,
              '--logo-color-light': colorLight,
            }}
          >
            <Component size={logoSize} className='logo-mono' />
            {Component.Color && (
              <Component.Color size={logoSize} className='logo-color' />
            )}
          </div>
        ))}
        <div
          className='logo-item-wrapper flex-shrink-0 flex items-center justify-center'
          style={{ width: logoSize + 40 }}
        >
          <Typography.Text
            className={`!text-lg sm:!text-xl md:!text-2xl font-bold ${isDark ? 'text-white/60' : 'text-black/50'}`}
          >
            30+
          </Typography.Text>
        </div>
      </>
    ),
    [logoSize, isDark],
  );

  const currentYear = new Date().getFullYear();

  /* ------------------------------------------------------------
     Color tokens
     ------------------------------------------------------------ */
  const bgBase = isDark ? '#050505' : '#f8fafc';
  const textPrimary = isDark ? 'text-white' : 'text-gray-900';
  const textSecondary = isDark ? 'text-white/35' : 'text-black/40';
  const textTertiary = isDark ? 'text-white/25' : 'text-black/25';
  const iconBg = isDark ? 'bg-white/[0.05]' : 'bg-black/[0.04]';
  const btnSecondaryBorder = isDark
    ? 'border-white/15 hover:border-white/30 hover:bg-white/5'
    : 'border-black/10 hover:border-black/20 hover:bg-black/5';
  const btnSecondaryText = isDark
    ? 'text-white/70 hover:text-white'
    : 'text-black/60 hover:text-black';

  return (
    <div className='w-full h-full overflow-hidden' style={fontStack}>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='w-full h-full overflow-hidden'>
          {/* ==================== Hero Section ==================== */}
          <div
            className={`w-full h-full relative overflow-hidden flex flex-col ${isDark ? 'home-hero-dark' : 'home-hero-light'}`}
          >
            {/* ---- Mesh Gradient Background ---- */}
            <div className='mesh-gradient-bg' aria-hidden='true' />

            {/* ---- Atmospheric Glow Orbs ---- */}
            <div className='hero-orb hero-orb-purple' aria-hidden='true' />
            <div className='hero-orb hero-orb-cyan' aria-hidden='true' />

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

            {/* ---- Grain texture overlay ---- */}
            <div className='grain-overlay' aria-hidden='true' />

            {/* ---- Soft purple glow behind text ---- */}
            <div className='hero-purple-glow' aria-hidden='true' />

            {/* ---- Center Content ---- */}
            <div className='relative z-10 flex-1 flex items-center justify-center px-4 pt-20 md:pt-24'>
              <div className='flex flex-col items-center justify-center text-center max-w-4xl mx-auto w-full'>
                {/* Brand name */}
                {/* <div className='animate-fade-in-up animate-fade-in-up-delay-1'>
                  <div className='gradient-text-apple text-4xl md:text-5xl lg:text-6xl font-black tracking-tighter mb-4 md:mb-5'>
                    New API
                  </div>
                </div> */}

                {/* Main title */}
                <h1
                  className='animate-fade-in-up animate-fade-in-up-delay-2 text-[1.75rem] md:text-[2.5rem] lg:text-[3rem] font-bold leading-[1.1] tracking-tight gradient-title-flow'
                >
                  <span className='block'>{t('统一的')}</span>
                  <span className='block'>{t('大模型接口网关')}</span>
                </h1>

                {/* Subtitle */}
                <p
                  className={`animate-fade-in-up animate-fade-in-up-delay-3 text-sm md:text-base mt-4 md:mt-5 max-w-lg font-medium tracking-wide ${textSecondary}`}
                >
                  {t('更好的价格，更好的稳定性，只需要将模型基址替换为：')}
                </p>

                {/* URL Block — Deep Blue Glass */}
                <div className='animate-fade-in-up animate-fade-in-up-delay-3 mt-7 md:mt-8 w-full max-w-md mx-auto px-2'>
                  <div
                    className='url-box-glass flex items-center gap-3 px-5 py-3.5 md:px-6 md:py-4 rounded-3xl transition-all duration-500'
                  >
                    <span
                      className={`text-sm md:text-[0.9rem] font-mono truncate shrink-0 tracking-wide ${isDark ? 'text-white/55' : 'text-black/55'}`}
                    >
                      {serverAddress}
                      <span className={isDark ? 'text-white/20' : 'text-black/20'}>
                        {currentEndpoint}
                      </span>
                    </span>

                    <span className='flex-1' />

                    <button
                      onClick={handleCopyBaseURL}
                      className='shrink-0 w-8 h-8 flex items-center justify-center rounded-full transition-all duration-500 hover:bg-white/10 group'
                      title={t('复制')}
                    >
                      <Copy
                        className={`w-3.5 h-3.5 transition-all duration-500 ${isDark ? 'text-white/25 group-hover:text-white/70' : 'text-black/25 group-hover:text-black/60'}`}
                        strokeWidth={1.5}
                      />
                    </button>
                  </div>
                </div>

                {/* Action buttons */}
                <div className='animate-fade-in-up animate-fade-in-up-delay-4 flex flex-row gap-3 md:gap-4 justify-center items-center mt-7 md:mt-9'>
                  <Link to='/console'>
                    <button
                      className={`hero-btn-press flex items-center gap-2 px-7 py-3 rounded-full text-sm font-semibold transition-all duration-500 ${isDark ? 'bg-white text-black' : 'bg-gray-900 text-white'}`}
                    >
                      <Play className='w-4 h-4' strokeWidth={1.5} />
                      {t('获取密钥')}
                    </button>
                  </Link>
                  {isDemoSiteMode && statusState?.status?.version ? (
                    <button
                      className={`hero-btn-press flex items-center gap-2 px-6 py-3 rounded-full text-sm font-medium backdrop-blur-xl border transition-all duration-500 ${btnSecondaryBorder} ${btnSecondaryText}`}
                      onClick={() =>
                        window.open(
                          'https://github.com/QuantumNous/new-api',
                          '_blank',
                        )
                      }
                    >
                      <Github className='w-4 h-4' strokeWidth={1.5} />
                      {statusState.status.version}
                    </button>
                  ) : (
                    docsLink && (
                      <button
                        className={`hero-btn-press flex items-center gap-2 px-6 py-3 rounded-full text-sm font-medium backdrop-blur-xl border transition-all duration-500 ${btnSecondaryBorder} ${btnSecondaryText}`}
                        onClick={() => window.open(docsLink, '_blank')}
                      >
                        <FileText className='w-4 h-4' strokeWidth={1.5} />
                        {t('文档')}
                      </button>
                    )
                  )}
                </div>

                {/* ==================== Bento Grid ==================== */}
                <div className='animate-fade-in-up animate-fade-in-up-delay-5 mt-14 md:mt-20 w-full max-w-[800px] mx-auto px-4'>
                  <div className='grid grid-cols-1 md:grid-cols-3 gap-3 md:gap-4'>
                    {/* Card 1 — Speed */}
                    <div
                      className='bento-card bento-card-blue flex flex-col items-start gap-3.5 p-5 md:p-6 rounded-[1.75rem]'
                    >
                      <div
                        className={`relative w-10 h-10 flex items-center justify-center rounded-2xl ${iconBg}`}
                      >
                        <span className='icon-glow-dot icon-glow-blue' aria-hidden='true' />
                        <Zap
                          className='relative z-10 w-5 h-5 bento-icon-blue'
                          strokeWidth={1.5}
                        />
                      </div>
                      <div className='text-left'>
                        <div className={`text-[0.95rem] md:text-[1.05rem] font-semibold tracking-tight ${textPrimary}`}>
                          {t('极速响应')}
                        </div>
                        <div
                          className={`text-[0.78rem] md:text-[0.82rem] mt-1 leading-relaxed ${textTertiary}`}
                        >
                          {t('毫秒级低延迟中转网关')}
                        </div>
                      </div>
                    </div>

                    {/* Card 2 — Reliability (slightly larger) */}
                    <div
                      className='bento-card bento-card-green flex flex-col items-start gap-3.5 p-5 md:p-7 rounded-[1.75rem] md:scale-[1.02] md:-my-0.5'
                    >
                      <div
                        className={`relative w-10 h-10 flex items-center justify-center rounded-2xl ${iconBg}`}
                      >
                        <span className='icon-glow-dot icon-glow-green' aria-hidden='true' />
                        <Shield
                          className='relative z-10 w-5 h-5 bento-icon-green'
                          strokeWidth={1.5}
                        />
                      </div>
                      <div className='text-left'>
                        <div className={`text-[0.95rem] md:text-[1.05rem] font-semibold tracking-tight ${textPrimary}`}>
                          {t('高可用性')}
                        </div>
                        <div
                          className={`text-[0.78rem] md:text-[0.82rem] mt-1 leading-relaxed ${textTertiary}`}
                        >
                          {t('99.9% 系统正常运行时间保证')}
                        </div>
                      </div>
                    </div>

                    {/* Card 3 — Cost */}
                    <div
                      className='bento-card bento-card-yellow flex flex-col items-start gap-3.5 p-5 md:p-6 rounded-[1.75rem]'
                    >
                      <div
                        className={`relative w-10 h-10 flex items-center justify-center rounded-2xl ${iconBg}`}
                      >
                        <span className='icon-glow-dot icon-glow-yellow' aria-hidden='true' />
                        <BarChart3
                          className='relative z-10 w-5 h-5 bento-icon-yellow'
                          strokeWidth={1.5}
                        />
                      </div>
                      <div className='text-left'>
                        <div className={`text-[0.95rem] md:text-[1.05rem] font-semibold tracking-tight ${textPrimary}`}>
                          {t('成本透明')}
                        </div>
                        <div
                          className={`text-[0.78rem] md:text-[0.82rem] mt-1 leading-relaxed ${textTertiary}`}
                        >
                          {t('统一多模型计费与实时账单')}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* ==================== Provider Logos ==================== */}
                <div className='animate-fade-in-up animate-fade-in-up-delay-5 mt-14 md:mt-20 w-full'>
                  <div className='flex items-center mb-6 md:mb-8 justify-center'>
                    <Text
                      className={`text-xs md:text-sm font-medium tracking-widest uppercase ${textTertiary}`}
                    >
                      {t('支持众多的大模型供应商')}
                    </Text>
                  </div>

                  {/* Infinite-scroll logo marquee */}
                  <div className='relative w-full overflow-hidden py-3'>
                    {/* Left fade mask */}
                    <div
                      className='absolute left-0 top-0 bottom-0 w-16 sm:w-20 md:w-32 z-10 pointer-events-none'
                      style={{
                        background: `linear-gradient(to right, ${bgBase} 0%, transparent 100%)`,
                      }}
                    />
                    {/* Right fade mask */}
                    <div
                      className='absolute right-0 top-0 bottom-0 w-16 sm:w-20 md:w-32 z-10 pointer-events-none'
                      style={{
                        background: `linear-gradient(to left, ${bgBase} 0%, transparent 100%)`,
                      }}
                    />

                    <div className='logo-marquee-track flex items-center'>
                      {marqueeLogos}
                      {marqueeLogos}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* ---- Inline Footer ---- */}
            <div className='relative z-10 flex-shrink-0 py-5 md:py-6 px-4 text-center'>
              <Typography.Text
                className={`text-xs ${textTertiary}`}
              >
                北京润德骥图科技有限公司 Copyright ©{currentYear} spacehpc.com 备案号：
                <a
                  href='https://beian.miit.gov.cn'
                  target='_blank'
                  rel='noopener noreferrer'
                  className={`${isDark ? 'text-white/35 hover:text-white/55' : 'text-black/30 hover:text-black/50'} transition-colors duration-500`}
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
