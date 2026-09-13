// Main application JavaScript
console.log('Local APK Builder initialized');

// Global error handler
window.addEventListener('error', function(e) {
    console.error('Application error:', e.message);
});

// HTMX configuration
document.addEventListener('htmx:configRequest', function(evt) {
    console.log('HTMX request:', evt.detail.path);
});

document.addEventListener('htmx:responseError', function(evt) {
    console.error('HTMX error:', evt.detail);
});
