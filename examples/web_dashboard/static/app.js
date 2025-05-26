class FlightDashboard {
    constructor() {
        console.log('🚀 FlightDashboard constructor called');
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.isConnected = false;
        
        this.initializeWebSocket();
        this.setupUI();
    }
    
    initializeWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        console.log('🔌 Attempting WebSocket connection to:', wsUrl);
        
        try {
            this.ws = new WebSocket(wsUrl);
            console.log('🔌 WebSocket object created');
            
            this.ws.onopen = () => {
                console.log('✅ WebSocket connected successfully');
                this.isConnected = true;
                this.reconnectDelay = 1000;
                this.updateConnectionStatus(true, 'Connected');
                this.hideErrorBanner();
            };
            
            this.ws.onmessage = (event) => {
                console.log('📨 WebSocket message received:', event.data.substring(0, 100) + '...');
                try {
                    const data = JSON.parse(event.data);
                    this.updateFlightData(data);
                } catch (error) {
                    console.error('❌ Error parsing WebSocket data:', error);
                }
            };
            
            this.ws.onclose = (event) => {
                console.log('🔌 WebSocket disconnected, code:', event.code, 'reason:', event.reason);
                this.isConnected = false;
                this.updateConnectionStatus(false, 'Disconnected');
                this.scheduleReconnect();
            };
            
            this.ws.onerror = (error) => {
                console.error('❌ WebSocket error:', error);
                this.updateConnectionStatus(false, 'Connection Error');
            };

        } catch (error) {
            console.error('❌ Failed to create WebSocket:', error);
            this.updateConnectionStatus(false, 'Connection Failed');
            this.scheduleReconnect();
        }
    }
    
    scheduleReconnect() {
        setTimeout(() => {
            if (!this.isConnected) {
                console.log(`🔄 Attempting to reconnect in ${this.reconnectDelay}ms...`);
                this.initializeWebSocket();
                this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
            }
        }, this.reconnectDelay);
    }
    
    updateConnectionStatus(connected, message) {
        const statusDot = document.getElementById('status-dot');
        const statusText = document.getElementById('status-text');
        
        if (connected) {
            statusDot.className = 'w-3 h-3 rounded-full bg-green-500 animate-pulse';
            statusText.textContent = message;
            statusText.className = 'text-sm text-green-400';
        } else {
            statusDot.className = 'w-3 h-3 rounded-full bg-red-500';
            statusText.textContent = message;
            statusText.className = 'text-sm text-red-400';
        }
    }
    
    updateFlightData(data) {
        if (!data.connected) {
            this.showErrorBanner(data.error || 'SimConnect not connected');
            this.clearFlightData();
            return;
        }
        
        this.hideErrorBanner();
        
        // Primary flight data
        this.updateElement('altitude', data.altitude.toFixed(0));
        this.updateElement('indicated-speed', data.indicatedSpeed.toFixed(0));
        this.updateElement('ground-speed', data.groundSpeed.toFixed(0));
        this.updateElement('vertical-speed', this.formatVerticalSpeed(data.verticalSpeed));
        this.updateElement('heading', data.headingMagnetic.toFixed(0));
        
        // Secondary flight data
        this.updateElement('bank-angle', this.formatAngle(data.bankAngle));
        this.updateElement('pitch-angle', this.formatAngle(data.pitchAngle));
        this.updateElement('engine-rpm', data.engineRPM.toFixed(0));
        this.updateElement('throttle-pos', data.throttlePos.toFixed(0));
          // Controls
        this.updateElement('gear-status', data.gearPosition > 0.5 ? 'DOWN' : 'UP');
        this.updateElement('flaps-pos', data.flapsPosition.toFixed(0));
        
        // Position
        this.updateElement('latitude', data.latitude.toFixed(4) + '°');
        this.updateElement('longitude', data.longitude.toFixed(4) + '°');
        
        // Weather data
        this.updateElement('ambient-temperature', data.ambientTemperature.toFixed(1) + '°C');
        this.updateElement('barometric-pressure', data.barometricPressure.toFixed(0) + ' mb');
        this.updateElement('wind-info', `${data.windSpeed.toFixed(0)} kts @ ${data.windDirection.toFixed(0)}°`);
        this.updateElement('visibility', (data.visibility / 1000).toFixed(1) + ' km');
        this.updateElement('cloud-coverage', data.cloudCoverage.toFixed(0) + '%');
        
        // Aircraft info
        this.updateElement('aircraft-title', data.aircraftTitle || 'Unknown');
        this.updateElement('on-ground', data.onGround ? 'YES' : 'NO');
        this.updateElement('parking-brake', data.parkingBrake ? 'SET' : 'OFF');
          // Simulation info
        this.updateElement('simulation-rate', data.simulationRate.toFixed(1) + 'x');
        this.updateElement('sim-paused', data.simPaused ? 'PAUSED' : 'RUNNING');
        this.updateElement('local-time', data.localTime || '--:--');
        this.updateElement('zulu-time', data.zuluTime || '--:--Z');
        
        // Statistics
        this.updateElement('data-count', data.dataCount.toLocaleString());
        this.updateElement('error-count', data.errorCount.toString());
        this.updateElement('error-count-large', data.errorCount.toString());
        this.updateElement('update-rate', data.updateRate.toFixed(1));
        this.updateElement('update-rate-large', data.updateRate.toFixed(1));
        
        // Last update
        const lastUpdate = new Date(data.lastUpdate);
        const now = new Date();
        const secondsAgo = Math.floor((now - lastUpdate) / 1000);
        this.updateElement('last-update', secondsAgo < 60 ? `${secondsAgo}s` : `${Math.floor(secondsAgo / 60)}m`);
    }
    
    updateElement(id, value) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    }
    
    formatAngle(angle) {
        const sign = angle >= 0 ? '+' : '';
        return `${sign}${angle.toFixed(1)}°`;
    }
    
    formatVerticalSpeed(vs) {
        const sign = vs >= 0 ? '+' : '';
        return `${sign}${vs.toFixed(0)}`;
    }
    
    showErrorBanner(message) {
        const banner = document.getElementById('error-banner');
        const messageEl = document.getElementById('error-message');
        
        if (banner && messageEl) {
            messageEl.textContent = message;
            banner.classList.remove('hidden');
        }
    }
    
    hideErrorBanner() {
        const banner = document.getElementById('error-banner');
        if (banner) {
            banner.classList.add('hidden');
        }
    }
      clearFlightData() {
        // Clear all flight data displays
        const fields = [            'altitude', 'indicated-speed', 'ground-speed', 'vertical-speed', 'heading',
            'bank-angle', 'pitch-angle', 'engine-rpm', 'throttle-pos',
            'gear-status', 'flaps-pos', 'latitude', 'longitude',
            'ambient-temperature', 'barometric-pressure', 'wind-info', 'visibility', 'cloud-coverage',
            'aircraft-title', 'on-ground', 'parking-brake',
            'simulation-rate', 'sim-paused', 'local-time', 'zulu-time',
            'data-count', 'error-count', 'error-count-large', 
            'update-rate', 'update-rate-large', 'last-update'
        ];
        
        fields.forEach(field => this.updateElement(field, '--'));
    }
    
    setupUI() {
        // Add any additional UI setup here
        console.log('🎨 Dashboard UI initialized');
        
        // Add keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 'r') {
                e.preventDefault();
                console.log('🔄 Manual refresh requested');
                location.reload();
            }
        });
    }
}

console.log('🎯 App.js loaded successfully, initializing dashboard...');

// Initialize dashboard when page loads
document.addEventListener('DOMContentLoaded', () => {
    console.log('🚀 DOM loaded, creating FlightDashboard instance...');
    window.flightDashboard = new FlightDashboard();
});

// Also initialize immediately if DOM is already loaded
if (document.readyState === 'loading') {
    console.log('⏳ DOM still loading, waiting for DOMContentLoaded...');
} else {
    console.log('🚀 DOM already loaded, creating FlightDashboard instance immediately...');
    window.flightDashboard = new FlightDashboard();
}

// Handle page visibility changes
document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
        console.log('📱 Page hidden - dashboard paused');
    } else {
        console.log('📱 Page visible - dashboard resumed');
        // Could implement reconnection logic here if needed
    }
});
