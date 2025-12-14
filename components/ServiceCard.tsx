import React from 'react';
import { Service, ServiceStatus } from '../types';
import { Settings, ExternalLink } from 'lucide-react';

interface ServiceCardProps {
  service: Service;
  onClick: (service: Service) => void;
}

const ServiceCard: React.FC<ServiceCardProps> = ({ service, onClick }) => {
  
  const getStatusColor = (status: ServiceStatus) => {
    switch (status) {
      case ServiceStatus.RUNNING: return 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.4)]';
      case ServiceStatus.STOPPED: return 'bg-rose-500 shadow-[0_0_10px_rgba(244,63,94,0.4)]';
      case ServiceStatus.ERROR: return 'bg-red-600 shadow-[0_0_10px_rgba(220,38,38,0.4)]';
      case ServiceStatus.MAINTENANCE: return 'bg-amber-500 shadow-[0_0_10px_rgba(245,158,11,0.4)]';
      default: return 'bg-slate-500';
    }
  };

  const handleOpenService = (e: React.MouseEvent) => {
    e.stopPropagation();
    const fullUrl = `${service.url}:${service.port}`;
    window.open(fullUrl, '_blank');
  };

  return (
    <div 
      className="group bg-slate-900/40 backdrop-blur-md border border-slate-800 hover:border-indigo-500/50 rounded-xl p-5 transition-all duration-300 hover:shadow-2xl hover:shadow-indigo-500/10 hover:-translate-y-1 cursor-pointer flex flex-col justify-between min-h-[140px] relative overflow-hidden"
      onClick={() => onClick(service)}
    >
      <div className="absolute top-3 right-3 flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300 translate-y-[-10px] group-hover:translate-y-0 z-10">
        <button
          onClick={handleOpenService}
          className="p-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg shadow-lg hover:shadow-indigo-500/25 transition-colors"
          title={`Open ${service.name}`}
        >
          <ExternalLink className="w-4 h-4" />
        </button>
        <div className="p-2 bg-slate-800/80 text-slate-400 rounded-lg backdrop-blur-sm">
          <Settings className="w-4 h-4" />
        </div>
      </div>

      <div className="flex items-center justify-between mb-4">
        <h3 className="text-xl font-bold text-white group-hover:text-indigo-200 transition-colors truncate pr-4">{service.name}</h3>
        <div className={`w-3 h-3 shrink-0 rounded-full ${getStatusColor(service.status)}`} title={service.status} />
      </div>

      <div className="mt-auto pt-4 border-t border-slate-800/50 flex items-center justify-between text-xs text-slate-500">
        <div className="flex flex-col">
           <span className="uppercase tracking-wider text-[10px] opacity-70">Uptime</span>
           <span className="font-mono text-slate-300">{service.uptime}</span>
        </div>
        <div className="flex flex-col items-end">
           <span className="uppercase tracking-wider text-[10px] opacity-70">Port</span>
           <span className="font-mono text-slate-300">{service.port}</span>
        </div>
      </div>
    </div>
  );
};

export default ServiceCard;