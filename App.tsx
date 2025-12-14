import React, { useState, useEffect, useCallback } from 'react';
import Login from './components/Login';
import ServiceCard from './components/ServiceCard';
import ConfigViewer from './components/ConfigViewer';
import HostStats from './components/HostStats';
import { Service } from './types';
import { getServices, checkAuth, logout } from './services/api';
import { LayoutGrid, LogOut, Search, Activity, Cpu, RefreshCw } from 'lucide-react';

const POLL_INTERVAL = 10000; // 10 seconds

const App: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [services, setServices] = useState<Service[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Check authentication status on mount
  useEffect(() => {
    checkAuth()
      .then(() => setIsAuthenticated(true))
      .catch(() => setIsAuthenticated(false))
      .finally(() => setIsCheckingAuth(false));
  }, []);

  // Fetch services
  const fetchServices = useCallback(async () => {
    if (!isAuthenticated) return;

    try {
      setIsLoading(true);
      const data = await getServices();
      setServices(data.services);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch services');
      // If unauthorized, redirect to login
      if (err.message?.includes('401') || err.message?.includes('Unauthorized')) {
        setIsAuthenticated(false);
      }
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated]);

  // Initial fetch and polling
  useEffect(() => {
    if (isAuthenticated) {
      fetchServices();
      const interval = setInterval(fetchServices, POLL_INTERVAL);
      return () => clearInterval(interval);
    }
  }, [isAuthenticated, fetchServices]);

  const handleLogout = async () => {
    try {
      await logout();
    } catch (err) {
      // Ignore errors, logout locally anyway
    }
    setIsAuthenticated(false);
    setServices([]);
  };

  const filteredServices = services.filter(service =>
    service.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const runningCount = services.filter(s => s.status === 'RUNNING').length;
  const totalCount = services.length;

  // Show loading while checking auth
  if (isCheckingAuth) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center">
        <div className="text-slate-400 flex items-center gap-2">
          <RefreshCw className="w-5 h-5 animate-spin" />
          <span>Loading...</span>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Login onLogin={() => setIsAuthenticated(true)} />;
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-200 font-sans selection:bg-indigo-500/30">

      {/* Navbar */}
      <header className="sticky top-0 z-40 w-full backdrop-blur-lg border-b border-slate-800/60 bg-slate-950/80">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16 gap-4">
            <div className="flex items-center gap-3 shrink-0">
              <div className="bg-indigo-600 p-2 rounded-lg shadow-lg shadow-indigo-500/20">
                <LayoutGrid className="w-5 h-5 text-white" />
              </div>
              <span className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-white to-slate-400 hidden sm:block">
                HomeLan
              </span>
            </div>

            {/* Search Bar - Moved to Header */}
            <div className="flex-1 max-w-md mx-auto">
              <div className="relative group">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Search className="h-4 w-4 text-slate-500 group-focus-within:text-indigo-400 transition-colors" />
                </div>
                <input
                  type="text"
                  className="block w-full pl-9 pr-3 py-2 border border-slate-800 rounded-lg leading-5 bg-slate-900/50 text-slate-300 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500/50 focus:border-indigo-500 sm:text-sm transition-all shadow-sm"
                  placeholder="Search services..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
            </div>

            <div className="flex items-center gap-6 shrink-0">
              <div className="hidden md:flex items-center gap-4 text-sm font-medium text-slate-400 bg-slate-900/50 py-1.5 px-4 rounded-full border border-slate-800">
                <div className="flex items-center gap-2">
                   <Activity className="w-4 h-4 text-emerald-400" />
                   <span>Running: <span className="text-white">{runningCount}</span></span>
                </div>
                <div className="w-px h-4 bg-slate-700"></div>
                <div className="flex items-center gap-2">
                   <Cpu className="w-4 h-4 text-indigo-400" />
                   <span>Total: <span className="text-white">{totalCount}</span></span>
                </div>
              </div>

              {/* Refresh button */}
              <button
                onClick={fetchServices}
                className={`p-2 text-slate-400 hover:text-indigo-400 transition-colors ${isLoading ? 'animate-spin' : ''}`}
                title="Refresh"
                disabled={isLoading}
              >
                <RefreshCw className="w-5 h-5" />
              </button>

              <button
                onClick={handleLogout}
                className="p-2 text-slate-400 hover:text-rose-400 transition-colors"
                title="Logout"
              >
                <LogOut className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">

        {/* Host Stats */}
        <HostStats />

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400">
            {error}
          </div>
        )}

        {/* Service Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {isLoading && services.length === 0 ? (
            // Loading skeletons
            Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="bg-slate-900/50 border border-slate-800 rounded-xl p-5 animate-pulse">
                <div className="h-6 bg-slate-800 rounded w-3/4 mb-4"></div>
                <div className="h-4 bg-slate-800 rounded w-1/2 mb-2"></div>
                <div className="h-4 bg-slate-800 rounded w-2/3"></div>
              </div>
            ))
          ) : filteredServices.length > 0 ? (
            filteredServices.map((service) => (
              <ServiceCard
                key={service.id}
                service={service}
                onClick={setSelectedService}
              />
            ))
          ) : services.length === 0 ? (
            <div className="col-span-full flex flex-col items-center justify-center py-20 text-slate-500">
               <div className="bg-slate-900/50 p-6 rounded-full mb-4">
                 <LayoutGrid className="w-10 h-10 opacity-50" />
               </div>
               <p className="text-lg">No services configured</p>
               <p className="text-sm mt-2">Add services to your config.yml file</p>
            </div>
          ) : (
            <div className="col-span-full flex flex-col items-center justify-center py-20 text-slate-500">
               <div className="bg-slate-900/50 p-6 rounded-full mb-4">
                 <Search className="w-10 h-10 opacity-50" />
               </div>
               <p className="text-lg">No services found matching "{searchTerm}"</p>
            </div>
          )}
        </div>
      </main>

      {/* Detail Modal */}
      {selectedService && (
        <ConfigViewer
          service={selectedService}
          onClose={() => setSelectedService(null)}
        />
      )}

      {/* Background ambient light */}
      <div className="fixed top-0 left-0 w-full h-full pointer-events-none -z-10 overflow-hidden">
         <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-indigo-900/10 rounded-full blur-[120px]"></div>
         <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-emerald-900/10 rounded-full blur-[120px]"></div>
      </div>

    </div>
  );
};

export default App;
