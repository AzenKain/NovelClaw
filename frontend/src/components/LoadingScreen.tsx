import React from 'react';
import { useTranslation } from 'react-i18next';
import { useAppStore } from '@/store/useAppStore';

export const LoadingScreen: React.FC = () => {
  const { t } = useTranslation();
  const loadingMessage = useAppStore((s) => s.loadingMessage);
  const appInfo = useAppStore((s) => s.appInfo);

  return (
    <div className="fixed inset-0 z-[9999] flex flex-col items-center justify-center bg-[#090d16]">
      {/* Ambient glow */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full bg-indigo-500/[0.04] blur-[120px]" />
        <div className="absolute top-1/3 left-1/3 w-[300px] h-[300px] rounded-full bg-indigo-400/[0.03] blur-[80px] animate-pulse" />
      </div>

      {/* Content */}
      <div className="relative flex flex-col items-center gap-8">
        {/* Logo / Brand */}
        <div className="flex flex-col items-center gap-3">
          <img src="/novelclaw-logo.png" alt="NovelClaw" className="w-16 h-16 rounded-2xl shadow-lg shadow-indigo-500/20" />
          <h1 className="text-2xl font-semibold tracking-tight text-white">
            {t('appName')}
          </h1>
          <p className="text-sm text-white/40">
            {t('tagline')}
          </p>
        </div>

        {/* Spinner */}
        <div className="relative">
          <div className="w-10 h-10 rounded-full border-2 border-indigo-500/20 border-t-indigo-500 animate-spin" />
          <div className="absolute inset-0 w-10 h-10 rounded-full border-2 border-transparent border-b-indigo-400/30 animate-spin" style={{ animationDirection: 'reverse', animationDuration: '1.5s' }} />
        </div>

        {/* Loading message */}
        <p className="text-sm text-white/50 animate-pulse min-h-[20px]">
          {loadingMessage ? t(loadingMessage) : t('common.loading')}
        </p>
      </div>

      {/* Version footer */}
      <div className="absolute bottom-6 text-xs text-white/20">
        {appInfo ? `v${appInfo.version}` : 'v1.0.0'}
      </div>
    </div>
  );
};
