import React, { useState, useEffect } from 'react';
import { Cpu, CircuitBoard, HardDrive, AlertCircle } from 'lucide-react';
import { getHostStats, HostStats as HostStatsType } from '../services/api';

const POLL_INTERVAL = 5000; // 5 seconds

const HostStats: React.FC = () => {
  const [stats, setStats] = useState<HostStatsType | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await getHostStats();
        setStats(data);
        setError(null);
      } catch (err: any) {
        setError(err.message || 'Failed to fetch stats');
      }
    };

    fetchStats();
    const interval = setInterval(fetchStats, POLL_INTERVAL);
    return () => clearInterval(interval);
  }, []);

  const getUsageColor = (percent: number) => {
    if (percent > 80) return 'bg-rose-500';
    if (percent > 60) return 'bg-amber-500';
    return 'bg-emerald-500';
  };

  const getTextColor = (percent: number) => {
    if (percent > 80) return 'text-rose-400';
    if (percent > 60) return 'text-amber-400';
    return 'text-emerald-400';
  };

  // Helper for progress bar
  const ProgressBar = ({ percent, colorClass }: { percent: number, colorClass: string }) => (
    <div className="w-full h-1.5 bg-slate-800 rounded-full mt-3 overflow-hidden">
      <div
        className={`h-full rounded-full transition-all duration-500 ${colorClass}`}
        style={{ width: `${percent}%` }}
      />
    </div>
  );

  // Show error state
  if (error) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="col-span-full bg-slate-900/40 backdrop-blur-md border border-slate-800 rounded-xl p-5">
          <div className="flex items-center gap-3 text-slate-400">
            <AlertCircle className="w-5 h-5 text-amber-500" />
            <span>Unable to load host stats: {error}</span>
          </div>
        </div>
      </div>
    );
  }

  // Show loading skeleton
  if (!stats) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-slate-900/40 backdrop-blur-md border border-slate-800 rounded-xl p-5 animate-pulse">
            <div className="h-6 bg-slate-800 rounded w-1/2 mb-4"></div>
            <div className="h-1.5 bg-slate-800 rounded w-full mt-3"></div>
            <div className="h-4 bg-slate-800 rounded w-1/3 mt-2"></div>
          </div>
        ))}
      </div>
    );
  }

  const cpuPercent = stats.cpu.usage;
  const memoryPercent = (stats.memory.usedGB / stats.memory.totalGB) * 100;
  const storagePercent = (stats.storage.usedGB / stats.storage.totalGB) * 100;

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      {/* CPU Card */}
      <div className="bg-slate-900/40 backdrop-blur-md border border-slate-800 rounded-xl p-5 hover:border-indigo-500/30 transition-all shadow-lg shadow-black/20 group">
        <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-3">
                <div className="p-2.5 bg-slate-800 rounded-lg group-hover:bg-indigo-500/10 transition-colors">
                    <Cpu className="w-5 h-5 text-indigo-400" />
                </div>
                <span className="text-sm font-medium text-slate-300">CPU Load</span>
            </div>
            <span className={`font-mono text-xl font-bold ${getTextColor(cpuPercent)}`}>{cpuPercent.toFixed(1)}%</span>
        </div>
        <ProgressBar percent={cpuPercent} colorClass={getUsageColor(cpuPercent)} />
        <p className="text-xs text-slate-500 mt-2 font-mono">{stats.cpu.cores} Cores / {stats.cpu.threads} Threads</p>
      </div>

      {/* RAM Card */}
      <div className="bg-slate-900/40 backdrop-blur-md border border-slate-800 rounded-xl p-5 hover:border-purple-500/30 transition-all shadow-lg shadow-black/20 group">
        <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-3">
                <div className="p-2.5 bg-slate-800 rounded-lg group-hover:bg-purple-500/10 transition-colors">
                    <CircuitBoard className="w-5 h-5 text-purple-400" />
                </div>
                <span className="text-sm font-medium text-slate-300">Memory</span>
            </div>
            <span className="font-mono text-xl font-bold text-white">
                {stats.memory.usedGB.toFixed(1)} <span className="text-sm font-normal text-slate-500 font-sans">GB</span>
            </span>
        </div>
        <ProgressBar percent={memoryPercent} colorClass="bg-purple-500" />
        <div className="flex justify-between mt-2">
           <p className="text-xs text-slate-500 font-mono">{stats.memory.totalGB.toFixed(0)} GB Total</p>
           <p className="text-xs text-slate-500">{memoryPercent.toFixed(0)}% Used</p>
        </div>
      </div>

      {/* Storage Card */}
      <div className="bg-slate-900/40 backdrop-blur-md border border-slate-800 rounded-xl p-5 hover:border-sky-500/30 transition-all shadow-lg shadow-black/20 group">
        <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-3">
                <div className="p-2.5 bg-slate-800 rounded-lg group-hover:bg-sky-500/10 transition-colors">
                    <HardDrive className="w-5 h-5 text-sky-400" />
                </div>
                <span className="text-sm font-medium text-slate-300">Storage</span>
            </div>
             <span className="font-mono text-xl font-bold text-white">
                {storagePercent.toFixed(0)}%
            </span>
        </div>
        <ProgressBar percent={storagePercent} colorClass="bg-sky-500" />
        <div className="flex justify-between mt-2 text-xs text-slate-500 font-mono">
             <span>Root Partition</span>
             <span>{stats.storage.usedGB.toFixed(0)}GB / {stats.storage.totalGB.toFixed(0)}GB</span>
        </div>
      </div>
    </div>
  );
};

export default HostStats;
