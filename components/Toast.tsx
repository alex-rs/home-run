import React, { useEffect } from 'react';
import { X, AlertCircle, CheckCircle, Info } from 'lucide-react';

export type ToastType = 'success' | 'error' | 'info';

interface ToastProps {
  message: string;
  type: ToastType;
  onClose: () => void;
}

const Toast: React.FC<ToastProps> = ({ message, type, onClose }) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose();
    }, 4000);
    return () => clearTimeout(timer);
  }, [onClose]);

  const getIcon = () => {
    switch (type) {
      case 'success': return <CheckCircle className="w-5 h-5 text-emerald-400" />;
      case 'error': return <AlertCircle className="w-5 h-5 text-rose-400" />;
      default: return <Info className="w-5 h-5 text-indigo-400" />;
    }
  };

  const getStyles = () => {
    switch (type) {
      case 'success': return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-200';
      case 'error': return 'border-rose-500/20 bg-rose-500/10 text-rose-200';
      default: return 'border-indigo-500/20 bg-indigo-500/10 text-indigo-200';
    }
  };

  return (
    <div className={`fixed top-6 right-6 z-[100] flex items-start gap-3 px-4 py-3 rounded-xl border backdrop-blur-xl shadow-2xl animate-in slide-in-from-top-2 fade-in duration-300 max-w-sm ${getStyles()}`}>
      <div className="mt-0.5 shrink-0">{getIcon()}</div>
      <div className="flex-1 text-sm leading-relaxed opacity-90">{message}</div>
      <button onClick={onClose} className="shrink-0 p-1 hover:bg-white/10 rounded-lg transition-colors -mr-1">
        <X className="w-4 h-4" />
      </button>
    </div>
  );
};

export default Toast;
