import React, { useEffect, useRef } from 'react';
import { AlertTriangle, AlertCircle, Info, Trash2, X, Check } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useConfirmStore } from '@/store/confirmStore';

export const ConfirmModal: React.FC = () => {
  const { t } = useTranslation();
  const { isOpen, options, close } = useConfirmStore();
  const confirmBtnRef = useRef<HTMLButtonElement>(null);

  const {
    title,
    message,
    confirmText,
    cancelText,
    variant = 'danger',
    icon,
    isAlert = false,
  } = options;

  // Focus confirm button when opened, handle Escape and Enter
  useEffect(() => {
    if (!isOpen) return;

    const timer = setTimeout(() => {
      confirmBtnRef.current?.focus();
    }, 50);

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        close(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, close]);

  if (!isOpen) return null;

  const defaultTitle = isAlert
    ? (variant === 'danger' ? t('common.error', 'Error') : t('common.warning', 'Warning'))
    : (variant === 'danger' ? t('common.confirmDelete', 'Confirm Deletion') : t('common.confirmTitle', 'Confirm Action'));

  const resolvedTitle = title || defaultTitle;
  const resolvedConfirmText = confirmText || (isAlert ? t('common.close', 'Understood') : (variant === 'danger' ? t('common.delete', 'Delete') : t('common.confirm', 'Confirm')));
  const resolvedCancelText = cancelText || t('common.cancel', 'Cancel');

  // Determine icon & color styling based on variant
  const renderIcon = () => {
    if (icon === 'trash' || (variant === 'danger' && !icon)) {
      return (
        <div className="w-11 h-11 rounded-2xl bg-rose-500/15 border border-rose-500/30 flex items-center justify-center text-rose-400 shrink-0 shadow-inner">
          <Trash2 className="w-5 h-5" />
        </div>
      );
    }
    if (variant === 'warning' || icon === 'warning') {
      return (
        <div className="w-11 h-11 rounded-2xl bg-amber-500/15 border border-amber-500/30 flex items-center justify-center text-amber-400 shrink-0 shadow-inner">
          <AlertCircle className="w-5 h-5" />
        </div>
      );
    }
    if (variant === 'success') {
      return (
        <div className="w-11 h-11 rounded-2xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400 shrink-0 shadow-inner">
          <Check className="w-5 h-5" />
        </div>
      );
    }
    return (
      <div className="w-11 h-11 rounded-2xl bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center text-indigo-400 shrink-0 shadow-inner">
        <Info className="w-5 h-5" />
      </div>
    );
  };

  const getConfirmBtnClasses = () => {
    if (variant === 'danger') {
      return 'bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white shadow-lg shadow-rose-950/40 border border-rose-400/30';
    }
    if (variant === 'warning') {
      return 'bg-amber-600 hover:bg-amber-500 active:bg-amber-700 text-white shadow-lg shadow-amber-950/40 border border-amber-400/30';
    }
    if (variant === 'success') {
      return 'bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white shadow-lg shadow-emerald-950/40 border border-emerald-400/30';
    }
    return 'bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white shadow-lg shadow-indigo-950/40 border border-indigo-400/30';
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-9999 bg-black/60 flex items-center justify-center p-4 select-none animate-in fade-in duration-150"
      onClick={() => close(false)}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="relative w-full max-w-md bg-[#0e1320] border border-white/10 rounded-2xl shadow-2xl p-5 sm:p-6 space-y-4 animate-in zoom-in-95 duration-150 ring-1 ring-white/5"
      >
        {/* Top Header */}
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center space-x-3 min-w-0">
            {renderIcon()}
            <div className="min-w-0">
              <h3 className="text-sm font-semibold text-white tracking-tight truncate">
                {resolvedTitle}
              </h3>
              <p className="text-[11px] text-slate-400 mt-0.5">
                {variant === 'danger' ? t('common.cannotUndo', 'This action cannot be undone') : t('common.systemConfirm', 'Please confirm this operation')}
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={() => close(false)}
            className="p-1 text-slate-400 hover:text-white rounded-lg hover:bg-white/8 transition"
            title={t('common.close', 'Close')}
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Modal Content / Message */}
        <div className="bg-[#080c14] border border-white/6 rounded-xl p-3.5 text-xs sm:text-[13px] text-slate-300 leading-relaxed font-sans select-text">
          {message}
        </div>

        {/* Footer Actions */}
        <div className="flex items-center justify-end space-x-2 pt-1">
          {!isAlert && (
            <button
              type="button"
              onClick={() => close(false)}
              className="px-3.5 py-1.5 rounded-xl text-xs font-medium text-slate-300 hover:text-white bg-white/4 hover:bg-white/8 border border-white/10 transition"
            >
              {resolvedCancelText}
            </button>
          )}
          <button
            ref={confirmBtnRef}
            type="button"
            onClick={() => close(true)}
            className={`px-4 py-1.5 rounded-xl text-xs font-semibold transition flex items-center space-x-1.5 ${getConfirmBtnClasses()}`}
          >
            {variant === 'danger' && <Trash2 className="w-3.5 h-3.5" />}
            {variant === 'warning' && <AlertTriangle className="w-3.5 h-3.5" />}
            {variant === 'success' && <Check className="w-3.5 h-3.5" />}
            <span>{resolvedConfirmText}</span>
          </button>
        </div>
      </div>
    </div>
  );
};
