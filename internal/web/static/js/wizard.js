// Navigation screen
function navScreen(existingConfig) {
    let config = { style: 'bottom', template: 'bottom-tabs', items: [] };
    if (existingConfig && existingConfig !== '{}') {
        try { config = JSON.parse(existingConfig); } catch(e) {}
    }
    return {
        style: config.style || 'bottom',
        template: config.template || 'bottom-tabs',
        items: config.items?.length
            ? config.items.map(i => ({ icon: i.icon || 'fa-house', url: i.url || '', animation: i.animation || 'spinner' }))
            : [],
        loadingAnimation: config.loadingAnimation || 'spinner',
        openDropdown: -1,
        search: '',
        icons: [
            'fa-house','fa-magnifying-glass','fa-user','fa-cart-shopping',
            'fa-heart','fa-bell','fa-envelope','fa-gear',
            'fa-camera','fa-location-dot','fa-phone','fa-globe',
            'fa-star','fa-bookmark','fa-comment','fa-share-nodes',
            'fa-download','fa-upload','fa-print','fa-trash',
            'fa-pen','fa-plus','fa-check','fa-bars',
            'fa-list','fa-table-cells','fa-moon','fa-sun',
            'fa-cloud','fa-wifi','fa-music','fa-film',
            'fa-image','fa-video','fa-microphone','fa-gamepad'
        ],
        get filteredIcons() {
            if (!this.search) return this.icons;
            return this.icons.filter(i => i.includes(this.search.toLowerCase()));
        },
        toggleDropdown(index) {
            this.openDropdown = this.openDropdown === index ? -1 : index;
            this.search = '';
        },
        selectIcon(index, icon) {
            if (index >= 0 && index < this.items.length) {
                this.items[index].icon = icon;
            }
            this.openDropdown = -1;
        },
        addItem() {
            if (this.items.length < 5) this.items.push({ icon: 'fa-house', url: '', animation: 'spinner' });
        },
        removeItem(index) {
            this.items.splice(index, 1);
        },
        async saveNavigation(projectId) {
            const data = { nav_style: this.style, template: this.template, items: this.items };
            try {
                const res = await fetch(`/api/projects/${projectId}/navigation`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (res.ok) window.location.href = `/wizard/6?project_id=${projectId}`;
                else alert('Save failed');
            } catch(e) { alert('Error: ' + e.message); }
        }
    };
}

// Logo upload handler (global)
function handleLogoSelect(event) {
    const file = event.target.files[0];
    if (!file) return;
    if (!file.type.startsWith('image/')) { alert('Please select an image'); return; }
    if (file.size > 5 * 1024 * 1024) { alert('Max 5MB'); return; }

    const reader = new FileReader();
    reader.onload = function(e) {
        const dataURL = e.target.result;
        document.getElementById('logo_path').value = dataURL;
        const img = document.getElementById('logo-image');
        const placeholder = document.getElementById('logo-placeholder');
        const btn = document.getElementById('logo-upload-btn');
        if (img) { img.src = dataURL; img.style.display = 'block'; }
        if (placeholder) placeholder.style.display = 'none';
        if (btn) { btn.textContent = '✓ Logo Selected'; btn.classList.add('has-logo'); }
    };
    reader.readAsDataURL(file);
}

// Color helper functions (global, used by Alpine)
function updateColorPreview() {
    const native = document.getElementById('primary_color');
    const text = document.getElementById('primary_color_text');
    if (native && text) text.value = native.value;
    // Update Alpine state if available
    const el = document.querySelector('[x-data]');
    if (el && el._x_dataStack) {
        const data = el._x_dataStack[0];
        if (data && 'primaryColor' in data) data.primaryColor = native.value;
    }
}

function updateColorPicker() {
    const native = document.getElementById('primary_color');
    const text = document.getElementById('primary_color_text');
    if (native && text && /^#[0-9A-Fa-f]{6}$/.test(text.value)) {
        native.value = text.value;
        const el = document.querySelector('[x-data]');
        if (el && el._x_dataStack) {
            const data = el._x_dataStack[0];
            if (data && 'primaryColor' in data) data.primaryColor = text.value;
        }
    }
}

// Branding screen Alpine component
function brandingScreen() {
    return {
        primaryColor: document.getElementById('primary_color')?.value || '#6366f1',
        setColor(color) {
            this.primaryColor = color;
            const native = document.getElementById('primary_color');
            const text = document.getElementById('primary_color_text');
            if (native) native.value = color;
            if (text) text.value = color;
        },
        updateColorPreview: updateColorPreview,
        updateColorPicker: updateColorPicker
    };
}

document.addEventListener('alpine:init', () => {
    Alpine.data('navScreen', navScreen);
    Alpine.data('brandingScreen', brandingScreen);
});

// Splash screen component
function splashScreen(bgColor, showLogo, loadingText, duration, logoPath) {
    return {
        bgColor: bgColor || '#6366f1',
        showLogo: showLogo === true || showLogo === 'true',
        loadingText: loadingText || 'Loading...',
        duration: parseInt(duration) || 1500,
        logoPath: logoPath || '',
        updatePreview() {
            this.bgColor = document.getElementById('splash_bg')?.value || this.bgColor;
        },
        updatePicker() {
            const text = document.getElementById('splash_bg_text');
            if (text && /^#[0-9A-Fa-f]{6}$/.test(text.value)) {
                this.bgColor = text.value;
            }
        }
    };
}
document.addEventListener('alpine:init', () => {
    Alpine.data('splashScreen', splashScreen);
});

// Features screen component
function featuresScreen(existingConfig) {
    let features = [];
    if (existingConfig && existingConfig !== '[]') {
        try {
            features = JSON.parse(existingConfig);
        } catch(e) {
            features = existingConfig.split(',').filter(Boolean);
        }
    }
    return {
        features: features,
        bridgeEnabled: false,
        bridgeUrl: '',
        bridgeAuthMode: 'none',
        bridgeNotificationMode: 'polling',
        bridgePollInterval: 30,
        toggle(feature) {
            const idx = this.features.indexOf(feature);
            if (idx > -1) {
                this.features.splice(idx, 1);
            } else {
                this.features.push(feature);
            }
        }
    };
}
document.addEventListener('alpine:init', () => {
    Alpine.data('featuresScreen', featuresScreen);
});

// GitHub connection status
function githubStatus() {
    return {
        connected: false,
        username: '',
        async loadStatus() {
            try {
                const res = await fetch('/api/github/status');
                const data = await res.json();
                this.connected = data.connected;
                this.username = data.username || '';
            } catch(e) {
                // silently fail
            }
        }
    };
}
document.addEventListener('alpine:init', () => {
    Alpine.data('githubStatus', githubStatus);
});
