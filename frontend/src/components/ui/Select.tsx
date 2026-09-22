import React from 'react';
import { ChevronDown } from 'lucide-react';

interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  containerClassName?: string;
  leftIcon?: React.ReactNode;
}

export const Select: React.FC<SelectProps> = ({
  className = '',
  containerClassName = '',
  leftIcon,
  children,
  ...props
}) => {
  const hasHeight = /\bh-(?:[0-9]|auto|full|fit|px)\b/.test(className);
  const heightClass = hasHeight ? '' : 'h-8';

  return (
    <div className={`relative inline-flex items-center min-w-0 max-w-full ${containerClassName}`}>
      {leftIcon && (
        <span className="absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none text-slate-400 z-10 flex items-center shrink-0">
          {leftIcon}
        </span>
      )}
      <select
        className={`w-full min-w-0 max-w-full truncate appearance-none -webkit-appearance-none bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 focus:ring-1 focus:ring-inset focus:ring-indigo-500/30 text-slate-200 rounded-md [&>option]:bg-[#111726] [&>option]:text-slate-200 ${
          leftIcon ? 'pl-8' : 'pl-3'
        } pr-8 text-xs font-medium cursor-pointer transition disabled:opacity-50 disabled:cursor-not-allowed leading-normal ${heightClass} ${className}`}
        {...props}
      >
        {children}
      </select>
      <ChevronDown className="w-3.5 h-3.5 text-slate-400 pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 shrink-0 z-10" />
    </div>
  );
};
