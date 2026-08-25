// Wizard-specific Alpine.js components
function wizard() {
    return {
        url: '',
        appName: '',
        primaryColor: '#3B82F6',
        logoPath: '',
        loading: false,
        
        init() {
            // Initialize from existing values if present
            const urlInput = document.getElementById('url');
            if (urlInput && urlInput.value) {
                this.url = urlInput.value;
            }
            
            const nameInput = document.getElementById('app_name');
            if (nameInput && nameInput.value) {
                this.appName = nameInput.value;
            }
            
            const colorInput = document.getElementById('primary_color');
            if (colorInput && colorInput.value) {
                this.primaryColor = colorInput.value;
            }
            
            const logoInput = document.getElementById('logo_path');
            if (logoInput && logoInput.value) {
                this.logoPath = logoInput.value;
            }
        },
        
        updateColorPicker() {
            const colorInput = document.getElementById('primary_color');
            const colorText = document.getElementById('primary_color_text');
            
            if (colorInput && colorText) {
                if (this.isValidHexColor(colorText.value)) {
                    colorInput.value = colorText.value;
                }
            }
        },
        
        updateColorPreview() {
            const colorInput = document.getElementById('primary_color');
            const colorText = document.getElementById('primary_color_text');
            
            if (colorInput && colorText) {
                colorText.value = colorInput.value;
                this.primaryColor = colorInput.value;
            }
        },
        
        isValidHexColor(color) {
            return /^#[0-9A-Fa-f]{6}$/.test(color);
        }
    };
}

// Register Alpine component
document.addEventListener('alpine:init', () => {
    Alpine.data('wizard', wizard);
});
