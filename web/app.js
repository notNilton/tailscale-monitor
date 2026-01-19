// Global state
let currentDevice = null;
let devices = [];
let refreshInterval = null;

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    loadDevices();
    setupEventListeners();
    startAutoRefresh();
});

// Setup event listeners
function setupEventListeners() {
    document.getElementById('refreshBtn').addEventListener('click', () => {
        if (currentDevice) {
            loadMetrics(currentDevice);
        }
        loadDevices();
    });

    document.getElementById('searchInput').addEventListener('input', (e) => {
        filterDevices(e.target.value);
    });

    document.getElementById('periodSelect').addEventListener('change', () => {
        // Future: load historical data
    });
}

// Load devices from /api/peers
async function loadDevices() {
    try {
        const response = await fetch('/api/peers');
        if (!response.ok) throw new Error('Failed to fetch peers');

        devices = await response.json();

        // Check health of each device
        await checkDevicesHealth();

        renderDeviceList(devices);
        updateDeviceCount(devices.filter(d => d.available).length);
    } catch (error) {
        console.error('Error loading devices:', error);
        showError('Failed to load devices');
    }
}

// Check health of all devices
async function checkDevicesHealth() {
    const healthChecks = devices.map(async (device) => {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 2000); // 2s timeout

            const response = await fetch(`http://${device.ip}:8080/health`, {
                signal: controller.signal
            });

            clearTimeout(timeoutId);
            device.available = response.ok;
        } catch (error) {
            device.available = false;
        }
    });

    await Promise.all(healthChecks);
}

// Render device list
function renderDeviceList(deviceList) {
    const container = document.getElementById('deviceList');

    if (deviceList.length === 0) {
        container.innerHTML = '<div class="loading">No devices found</div>';
        return;
    }

    container.innerHTML = deviceList.map(device => {
        const isAvailable = device.available === true;
        const isActive = currentDevice?.ip === device.ip;
        const clickHandler = isAvailable ? `onclick="selectDevice('${device.ip}', '${device.hostname}')"` : '';
        const cursorStyle = isAvailable ? 'cursor: pointer;' : 'cursor: not-allowed; opacity: 0.5;';

        return `
            <div class="device-item ${!isAvailable ? 'offline' : ''} ${isActive ? 'active' : ''}" 
                 data-ip="${device.ip}" 
                 data-hostname="${device.hostname}"
                 ${clickHandler}
                 style="${cursorStyle}">
                <div class="device-item-header">
                    <span class="status-dot" style="background: ${isAvailable ? 'var(--success)' : 'var(--gray)'}"></span>
                    <span class="device-item-name">${device.hostname}</span>
                </div>
                <div class="device-item-ip">${device.ip}</div>
            </div>
        `;
    }).join('');
}

// Filter devices
function filterDevices(query) {
    const filtered = devices.filter(device =>
        device.hostname.toLowerCase().includes(query.toLowerCase()) ||
        device.ip.includes(query)
    );
    renderDeviceList(filtered);
}

// Update device count badge
function updateDeviceCount(count) {
    document.getElementById('deviceCount').textContent = count;
}

// Select device
function selectDevice(ip, hostname) {
    currentDevice = { ip, hostname };

    // Update UI
    renderDeviceList(devices);
    document.getElementById('welcomeScreen').style.display = 'none';
    document.getElementById('metricsPanel').style.display = 'block';
    document.getElementById('errorMessage').style.display = 'none';

    // Update device header
    document.getElementById('deviceName').textContent = hostname;
    document.getElementById('deviceIP').textContent = ip;

    // Load metrics
    loadMetrics(currentDevice);
}

// Load metrics from device
async function loadMetrics(device) {
    try {
        const response = await fetch(`http://${device.ip}:8080/status`);
        if (!response.ok) throw new Error('Failed to fetch metrics');

        const metrics = await response.json();
        renderMetrics(metrics);
        hideError();
    } catch (error) {
        console.error('Error loading metrics:', error);
        showError(`Failed to load metrics from ${device.hostname}`);
    }
}

// Render metrics
function renderMetrics(metrics) {
    // CPU
    document.getElementById('cpuUsage').textContent = `${metrics.cpu.usage_percent.toFixed(1)}%`;
    document.getElementById('cpuCores').textContent = metrics.cpu.cores;
    if (metrics.cpu.load_average && metrics.cpu.load_average.length >= 3) {
        document.getElementById('cpuLoad').textContent =
            metrics.cpu.load_average.map(l => l.toFixed(2)).join(', ');
    }

    // Memory
    document.getElementById('memUsage').textContent = `${metrics.memory.usage_percent.toFixed(1)}%`;
    document.getElementById('memUsed').textContent =
        `${formatBytes(metrics.memory.used_bytes)} / ${formatBytes(metrics.memory.total_bytes)}`;
    document.getElementById('memSwap').textContent =
        `${formatBytes(metrics.memory.swap_used_bytes)} / ${formatBytes(metrics.memory.swap_total_bytes)}`;

    // Disk
    renderDiskMetrics(metrics.disk);

    // Network
    document.getElementById('netRecv').textContent = formatBytes(metrics.network.bytes_recv);
    document.getElementById('netSent').textContent = formatBytes(metrics.network.bytes_sent);

    // System
    document.getElementById('sysOS').textContent =
        `${metrics.system.platform} ${metrics.system.platform_version}`;
    document.getElementById('sysUptime').textContent = formatUptime(metrics.system.uptime_seconds);
    document.getElementById('sysKernel').textContent = metrics.system.kernel_version;
}

// Render disk metrics
function renderDiskMetrics(disks) {
    const container = document.getElementById('diskMetrics');

    if (!disks || disks.length === 0) {
        container.innerHTML = '<div class="loading-small">No disk data</div>';
        return;
    }

    container.innerHTML = disks.map(disk => `
        <div class="disk-item">
            <div class="disk-item-header">
                <span class="disk-mount">${disk.mount_point}</span>
                <span class="disk-usage">${disk.usage_percent.toFixed(1)}%</span>
            </div>
            <div class="disk-bar">
                <div class="disk-bar-fill" style="width: ${disk.usage_percent}%"></div>
            </div>
            <div style="font-size: 0.85rem; color: var(--gray); margin-top: 4px;">
                ${formatBytes(disk.used_bytes)} / ${formatBytes(disk.total_bytes)}
            </div>
        </div>
    `).join('');
}

// Format bytes
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

// Format uptime
function formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (days > 0) {
        return `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
        return `${hours}h ${minutes}m`;
    } else {
        return `${minutes}m`;
    }
}

// Show error
function showError(message) {
    document.getElementById('errorMessage').style.display = 'flex';
    document.getElementById('errorText').textContent = message;
}

// Hide error
function hideError() {
    document.getElementById('errorMessage').style.display = 'none';
}

// Auto refresh
function startAutoRefresh() {
    // Refresh every 30 seconds
    refreshInterval = setInterval(() => {
        if (currentDevice) {
            loadMetrics(currentDevice);
        }
        loadDevices();
    }, 30000);
}

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (refreshInterval) {
        clearInterval(refreshInterval);
    }
});
