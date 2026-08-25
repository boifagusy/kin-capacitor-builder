function wizard() {
    return {
        primaryColor: '#6366f1',
        
        init() {
            const colorInput = document.getElementById('primary_color');
            if (colorInput && colorInput.value) this.primaryColor = colorInput.value;
        },
        
        setColor(color) {
            this.primaryColor = color;
            const colorInput = document.getElementById('primary_color');
            const colorText = document.getElementById('primary_color_text');
            if (colorInput) colorInput.value = color;
            if (colorText) colorText.value = color;
        },
        
        updateColorPicker() {
            const colorInput = document.getElementById('primary_color');
            const colorText = document.getElementById('primary_color_text');
            if (colorInput && colorText) {
                if (this.isValidHexColor(colorText.value)) {
                    colorInput.value = colorText.value;
                    this.primaryColor = colorText.value;
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

// Logo upload handling
function handleLogoSelect(event) {
    const file = event.target.files[0];
    if (!file) return;
    
    // Check file type
    if (!file.type.startsWith('image/')) {
        alert('Please select an image file');
        return;
    }
    
    // Check file size (max 5MB)
    if (file.size > 5 * 1024 * 1024) {
        alert('Image must be less than 5MB');
        return;
    }
    
    // Read file as data URL
    const reader = new FileReader();
    reader.onload = function(e) {
        const dataURL = e.target.result;
        
        // Store in hidden input
        const logoPathInput = document.getElementById('logo_path');
        logoPathInput.value = dataURL;
        
        // Show preview
        const previewImg = document.getElementById('logo-image');
        const placeholder = document.getElementById('logo-placeholder');
        const uploadBtn = document.getElementById('logo-upload-btn');
        
        previewImg.src = dataURL;
        previewImg.style.display = 'block';
        placeholder.style.display = 'none';
        
        uploadBtn.textContent = '✓ Logo Selected';
        uploadBtn.classList.add('has-logo');
    };
    reader.readAsDataURL(file);
}

document.addEventListener('alpine:init', () => {
    Alpine.data('wizard', wizard);
});
