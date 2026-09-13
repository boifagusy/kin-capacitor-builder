package runtime

const testHTML = `<!DOCTYPE html>
<html>
<head><title>Probe</title></head>
<body>
<h1>Probe</h1>
<script>
async function run() {
  try {
    const h = await fetch('/api/health').then(r => r.json());
    console.log('PROBE_WEBVIEW: api_health=' + JSON.stringify(h));
  } catch (e) {
    console.log('PROBE_WEBVIEW: api_health_error=' + e.message);
  }
  try {
    const r = await fetch('/api/notes', { method: 'POST', body: JSON.stringify({title:'hello', body:'world'}) });
    console.log('PROBE_WEBVIEW: note_create_status=' + r.status);
  } catch (e) {
    console.log('PROBE_WEBVIEW: note_create_error=' + e.message);
  }
  try {
    const r = await fetch('/api/notes', { method: 'POST', body: JSON.stringify({title:'hello2', body:'world2'}) });
    console.log('PROBE_WEBVIEW: note_create2_status=' + r.status);
  } catch (e) {
    console.log('PROBE_WEBVIEW: note_create2_error=' + e.message);
  }
  console.log('PROBE_WEBVIEW: done');
}
run();
</script>
</body>
</html>
`
