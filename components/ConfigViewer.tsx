import React, { useState, useEffect } from 'react';
import { Service, ServiceConfig } from '../types';
import SimpleHighlighter from './SyntaxHighlighter';
import { X, FileCode, Cpu, Terminal, Copy, Check, ExternalLink, BarChart3, Settings, FileText, Clock, RefreshCw } from 'lucide-react';
import { analyzeConfiguration } from '../services/geminiService';
import { getServiceConfig } from '../services/api';
import Toast, { ToastType } from './Toast';

interface ConfigViewerProps {
  service: Service;
  onClose: () => void;
}

type TabMode = 'config' | 'metrics';
type ConfigSubMode = 'code' | 'analysis';

// Helper component for stacked-style bar charts
const ResourceChart: React.FC<{
  data: number[];
  max: number;
  colorClass: string;
  label: string;
  unit: string;
  average: number;
}> = ({ data, max, colorClass, label, unit, average }) => {
  return (
    <div className="bg-slate-800/50 rounded-xl p-5 border border-slate-700/50">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-slate-200 font-medium text-sm flex items-center gap-2">
            {label}
            <span className="text-xs px-2 py-0.5 rounded-full bg-slate-700 text-slate-400">Last Hour</span>
          </h3>
        </div>
        <div className="text-right">
          <div className={`text-2xl font-bold ${colorClass.replace('bg-', 'text-')}`}>
            {average.toFixed(1)} <span className="text-sm font-normal text-slate-500">{unit}</span>
          </div>
        </div>
      </div>

      {/* Chart Area */}
      <div className="h-40 flex items-end justify-between gap-1 mt-4">
        {data.map((value, i) => (
          <div key={i} className="relative w-full h-full flex items-end group">
            {/* Tooltip */}
            <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block z-10">
              <div className="bg-slate-900 text-xs text-white px-2 py-1 rounded border border-slate-700 whitespace-nowrap">
                {value.toFixed(1)}{unit}
              </div>
            </div>
            {/* Bar - "Stacked" visual by having a background track */}
            <div className="w-full bg-slate-700/30 rounded-t-sm h-full relative overflow-hidden">
                <div
                  className={`w-full absolute bottom-0 transition-all duration-500 ${colorClass}`}
                  style={{ height: `${(value / max) * 100}%` }}
                ></div>
            </div>
          </div>
        ))}
      </div>
      <div className="flex justify-between mt-2 text-[10px] text-slate-500 font-mono uppercase">
        <span>60m ago</span>
        <span>Now</span>
      </div>
    </div>
  );
};

const ConfigViewer: React.FC<ConfigViewerProps> = ({ service, onClose }) => {
  const [activeTab, setActiveTab] = useState<TabMode>('config');
  const [selectedFileIndex, setSelectedFileIndex] = useState(0);
  const [subMode, setSubMode] = useState<ConfigSubMode>('code');

  const [isAnalyzing, setIsAnalyzing] = useState(false);
  // Store analysis results per file index
  const [analysisResults, setAnalysisResults] = useState<Record<number, string>>({});

  // Store loaded config content per file index
  const [loadedConfigs, setLoadedConfigs] = useState<Record<number, ServiceConfig>>({});
  const [isLoadingConfig, setIsLoadingConfig] = useState(false);
  const [configError, setConfigError] = useState<string | null>(null);

  const [copied, setCopied] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: ToastType } | null>(null);

  // Mock metrics data state
  const [metricsData, setMetricsData] = useState<{
    cpu: number[];
    memory: number[];
  }>({ cpu: [], memory: [] });

  const activeConfig = service.configs?.[selectedFileIndex];

  // Load config content when file is selected
  useEffect(() => {
    const loadConfig = async () => {
      // If already loaded or no configs, skip
      if (loadedConfigs[selectedFileIndex] || !service.configs?.length) {
        return;
      }

      // If content is already present (from API that includes content), use it
      if (activeConfig?.content) {
        setLoadedConfigs(prev => ({ ...prev, [selectedFileIndex]: activeConfig }));
        return;
      }

      setIsLoadingConfig(true);
      setConfigError(null);

      try {
        const config = await getServiceConfig(service.id, selectedFileIndex);
        setLoadedConfigs(prev => ({ ...prev, [selectedFileIndex]: config }));
      } catch (err: any) {
        setConfigError(err.message || 'Failed to load config');
      } finally {
        setIsLoadingConfig(false);
      }
    };

    if (activeTab === 'config') {
      loadConfig();
    }
  }, [selectedFileIndex, activeTab, service.id, service.configs?.length, activeConfig?.content, loadedConfigs]);

  // Generate mock historical data when service changes
  useEffect(() => {
    const dataPoints = 24; // e.g., every 2.5 mins for an hour

    // Generate CPU data with some variance around the current usage
    const cpuHistory = Array.from({ length: dataPoints }, () => {
      const variance = Math.random() * 10 - 5;
      return Math.max(0, Math.min(100, service.cpuUsage + variance));
    });

    // Generate Memory data (MB)
    const memHistory = Array.from({ length: dataPoints }, () => {
      const variance = Math.random() * 200 - 100;
      return Math.max(0, service.memoryUsage + variance);
    });

    setMetricsData({ cpu: cpuHistory, memory: memHistory });
  }, [service]);

  // Get the config content to display (from loaded or original)
  const displayConfig = loadedConfigs[selectedFileIndex] || activeConfig;
  const hasContent = displayConfig?.content && displayConfig.content.length > 0;

  const handleAnalyze = async () => {
    setSubMode('analysis');

    // If we already have a result for this specific file, don't re-fetch
    if (analysisResults[selectedFileIndex]) return;

    // Need content to analyze
    if (!hasContent) {
      setToast({ message: 'Config content not loaded yet', type: 'error' });
      return;
    }

    try {
      setIsAnalyzing(true);
      const result = await analyzeConfiguration(displayConfig.content, displayConfig.type);
      setAnalysisResults(prev => ({ ...prev, [selectedFileIndex]: result }));
    } catch (error: any) {
      setToast({ message: error.message, type: 'error' });
    } finally {
      setIsAnalyzing(false);
    }
  };

  const handleCopy = () => {
    if (!hasContent) return;
    navigator.clipboard.writeText(displayConfig.content);
    setCopied(true);
    setToast({ message: 'Configuration copied to clipboard', type: 'success' });
    setTimeout(() => setCopied(false), 2000);
  };

  const handleOpenService = () => {
    window.open(service.url, '_blank');
  };

  // Safe navigation between files
  const handleFileSelect = (index: number) => {
    setSelectedFileIndex(index);
    setSubMode('code'); // Reset to code view when switching files
    setConfigError(null);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 md:p-8 bg-black/60 backdrop-blur-sm">
      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}

      <div className="bg-slate-900 w-full max-w-5xl h-[90vh] rounded-2xl border border-slate-700 shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in duration-200">

        {/* Main Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800 bg-slate-950">
          <div className="flex items-center gap-4">
            <div className="p-2.5 bg-slate-900 border border-slate-800 rounded-xl">
              <Settings className="w-6 h-6 text-indigo-400" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white flex items-center gap-2">
                {service.name}
                {service.host && service.host !== 'local' && (
                  <span className="text-xs px-2 py-0.5 bg-slate-800 rounded text-slate-400">{service.host}</span>
                )}
                <button
                  onClick={handleOpenService}
                  className="p-1 text-slate-400 hover:text-indigo-400 transition-colors"
                  title="Open Service"
                >
                  <ExternalLink className="w-4 h-4" />
                </button>
              </h2>
              <div className="flex items-center gap-3 text-xs text-slate-400 mt-1">
                <span className={`px-2 py-0.5 rounded-full bg-slate-800/50 border border-slate-700 ${service.status === 'RUNNING' ? 'text-emerald-400 border-emerald-500/20' : 'text-slate-400'}`}>
                  {service.status}
                </span>
                <span className="font-mono">{service.url}:{service.port}</span>
              </div>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-slate-800 rounded-lg transition-colors text-slate-400 hover:text-white"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Top Tab Navigation */}
        <div className="flex items-center px-6 border-b border-slate-800 bg-slate-900">
          <button
            onClick={() => setActiveTab('config')}
            className={`flex items-center gap-2 px-4 py-4 text-sm font-medium border-b-2 transition-all ${
              activeTab === 'config'
                ? 'border-indigo-500 text-white'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <FileCode className="w-4 h-4" />
            Configuration
            <span className="ml-1 px-1.5 py-0.5 rounded-md bg-slate-800 text-[10px] text-slate-400">
              {service.configs?.length ?? 0}
            </span>
          </button>
          <button
            onClick={() => setActiveTab('metrics')}
            className={`flex items-center gap-2 px-4 py-4 text-sm font-medium border-b-2 transition-all ${
              activeTab === 'metrics'
                ? 'border-indigo-500 text-white'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <BarChart3 className="w-4 h-4" />
            Metrics
          </button>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-hidden relative flex bg-[#0d1117]">

          {/* CONFIGURATION VIEW */}
          {activeTab === 'config' && (
            <>
              {/* File Sidebar (only if > 1 config) */}
              {(service.configs?.length ?? 0) > 0 && (
                <div className={`${(service.configs?.length ?? 0) > 1 ? 'w-64 border-r border-slate-800' : 'w-0 hidden'} bg-slate-900/50 flex flex-col`}>
                  <div className="p-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                    Files
                  </div>
                  <div className="flex-1 overflow-y-auto custom-scrollbar px-2">
                    {service.configs?.map((conf, idx) => (
                      <button
                        key={idx}
                        onClick={() => handleFileSelect(idx)}
                        className={`w-full text-left px-3 py-2.5 rounded-lg mb-1 text-sm flex items-center gap-2 transition-colors ${
                          selectedFileIndex === idx
                            ? 'bg-indigo-600/10 text-indigo-300 border border-indigo-500/20'
                            : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200 border border-transparent'
                        }`}
                      >
                        <FileText className="w-4 h-4 shrink-0 opacity-70" />
                        <div className="flex-1 min-w-0">
                           <div className="truncate" title={conf.path}>{conf.path.split('/').pop()}</div>
                           <div className="text-[10px] text-slate-500 truncate">{conf.lastEdited}</div>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Editor Area */}
              <div className="flex-1 flex flex-col min-w-0 bg-[#0d1117]">
                {!service.configs?.length ? (
                  <div className="flex-1 flex flex-col items-center justify-center text-slate-500">
                    <FileCode className="w-12 h-12 opacity-20 mb-4" />
                    <p>No configuration files defined for this service</p>
                  </div>
                ) : (
                  <>
                    {/* Editor Toolbar */}
                    <div className="flex items-center gap-2 px-6 py-2 border-b border-slate-800 bg-slate-900/30">
                      <div className="flex items-center gap-1 bg-slate-800/50 p-1 rounded-lg">
                        <button
                          onClick={() => setSubMode('code')}
                          className={`px-3 py-1.5 text-xs font-medium rounded-md transition-all ${
                            subMode === 'code'
                              ? 'bg-slate-700 text-white shadow-sm'
                              : 'text-slate-400 hover:text-white'
                          }`}
                        >
                          Code
                        </button>
                        <button
                          onClick={handleAnalyze}
                          disabled={!hasContent}
                          className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all ${
                            subMode === 'analysis'
                              ? 'bg-indigo-500/20 text-indigo-300'
                              : 'text-slate-400 hover:text-indigo-300 disabled:opacity-50 disabled:cursor-not-allowed'
                          }`}
                        >
                          <Cpu className="w-3 h-3" />
                          AI Analysis
                        </button>
                      </div>

                      <div className="flex-1 text-center">
                        <span className="text-xs font-mono text-slate-500 block">{activeConfig?.path}</span>
                      </div>

                      <button
                        onClick={handleCopy}
                        disabled={!hasContent}
                        className="p-1.5 text-slate-400 hover:text-white transition-colors hover:bg-slate-800 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                        title="Copy Config"
                      >
                        {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                      </button>
                    </div>

                    {/* Main Content (Code or Analysis) */}
                    <div className="flex-1 overflow-auto custom-scrollbar p-6">
                      {isLoadingConfig ? (
                        <div className="h-64 flex flex-col items-center justify-center text-slate-400">
                          <RefreshCw className="w-8 h-8 mb-4 animate-spin text-indigo-500" />
                          <p>Loading configuration...</p>
                        </div>
                      ) : configError ? (
                        <div className="h-64 flex flex-col items-center justify-center text-red-400">
                          <FileCode className="w-12 h-12 opacity-50 mb-4" />
                          <p>{configError}</p>
                        </div>
                      ) : subMode === 'code' ? (
                        hasContent ? (
                          <SimpleHighlighter code={displayConfig.content} language={displayConfig.type} />
                        ) : (
                          <div className="h-64 flex flex-col items-center justify-center text-slate-500">
                            <FileCode className="w-12 h-12 opacity-20 mb-4" />
                            <p>No content available</p>
                          </div>
                        )
                      ) : (
                        <div className="max-w-3xl mx-auto">
                          {isAnalyzing ? (
                            <div className="h-64 flex flex-col items-center justify-center text-slate-400 animate-pulse">
                              <Cpu className="w-12 h-12 mb-4 text-indigo-500 animate-spin-slow" />
                              <p>Analyzing {activeConfig?.path?.split('/').pop() ?? 'config'} with Gemini...</p>
                            </div>
                          ) : analysisResults[selectedFileIndex] ? (
                            <div className="prose prose-invert prose-indigo max-w-none">
                              <div className="whitespace-pre-wrap font-sans text-sm text-slate-300 leading-relaxed">
                                {analysisResults[selectedFileIndex].split('\n').map((line, i) => {
                                  if (line.startsWith('# ')) return <h1 key={i} className="text-2xl font-bold text-white mb-4 mt-6 pb-2 border-b border-slate-800">{line.replace('# ', '')}</h1>
                                  if (line.startsWith('## ')) return <h2 key={i} className="text-xl font-bold text-indigo-200 mb-3 mt-5">{line.replace('## ', '')}</h2>
                                  if (line.startsWith('### ')) return <h3 key={i} className="text-lg font-bold text-white mb-2 mt-4">{line.replace('### ', '')}</h3>
                                  if (line.startsWith('- ')) return <li key={i} className="ml-4 mb-1 text-slate-300">{line.replace('- ', '')}</li>
                                  if (line.trim() === '') return <br key={i} />;
                                  return <p key={i} className="mb-2">{line}</p>
                                })}
                              </div>
                            </div>
                          ) : (
                             <div className="h-full flex flex-col items-center justify-center text-slate-500 mt-20">
                                <Cpu className="w-12 h-12 opacity-20 mb-4" />
                                <p className="max-w-xs text-center">Select "AI Analysis" to scan this file for security risks and best practices.</p>
                             </div>
                          )}
                        </div>
                      )}
                    </div>

                    {/* Footer Info */}
                    <div className="bg-slate-900 border-t border-slate-800 px-6 py-2 flex items-center justify-between text-[10px] text-slate-500 font-mono uppercase tracking-wider">
                      <div className="flex items-center gap-4">
                         <span>{displayConfig?.type || 'UNKNOWN'}</span>
                         <span>{hasContent ? `${displayConfig.content.length} BYTES` : '-- BYTES'}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Clock className="w-3 h-3" />
                        <span>Edited: {displayConfig?.lastEdited || 'Unknown'}</span>
                      </div>
                    </div>
                  </>
                )}
              </div>
            </>
          )}

          {/* METRICS VIEW */}
          {activeTab === 'metrics' && (
            <div className="flex-1 p-8 overflow-y-auto custom-scrollbar bg-slate-900">
              <div className="max-w-4xl mx-auto space-y-8">
                <div className="mb-8">
                  <h2 className="text-2xl font-bold text-white mb-2">Resource Usage</h2>
                  <p className="text-slate-400">Real-time performance metrics for the last hour.</p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <ResourceChart
                    label="CPU Usage"
                    unit="%"
                    data={metricsData.cpu}
                    max={100}
                    colorClass="bg-indigo-500"
                    average={service.cpuUsage}
                  />

                  <ResourceChart
                    label="Memory Usage"
                    unit="MB"
                    data={metricsData.memory}
                    max={Math.max(...metricsData.memory, 1) * 1.2}
                    colorClass="bg-emerald-500"
                    average={service.memoryUsage}
                  />
                </div>

                {/* Additional Info Box */}
                <div className="bg-slate-800/30 border border-slate-800 rounded-xl p-6 flex items-start gap-4">
                  <div className="p-3 bg-slate-800 rounded-lg">
                    <Terminal className="w-6 h-6 text-slate-400" />
                  </div>
                  <div>
                    <h4 className="text-white font-medium mb-1">Service Health</h4>
                    <p className="text-sm text-slate-400 mb-3">
                      The service has been running for <span className="text-white font-mono">{service.uptime || 'Unknown'}</span>.
                    </p>
                    <div className="flex gap-2 flex-wrap">
                       <span className="text-[10px] px-2 py-1 bg-slate-800 border border-slate-700 rounded text-slate-400">Port: {service.port}</span>
                       <span className="text-[10px] px-2 py-1 bg-slate-800 border border-slate-700 rounded text-slate-400">Status: {service.status}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ConfigViewer;
