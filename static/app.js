const API = 'http://localhost:3000';

// ── Toast notifications ──────────────────────────────────────────────────────
function toast(msg, type = 'success') {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.className = `show ${type}`;
  setTimeout(() => el.className = '', 3000);
}

// ── Load all devices ─────────────────────────────────────────────────────────
async function loadDevices() {
  try {
    const res = await fetch(`${API}/devices`);
    const data = await res.json();
    const devices = data.devices || [];

    document.getElementById('device-count').textContent = `(${devices.length})`;

    const tbody = document.getElementById('device-table-body');

    if (devices.length === 0) {
      tbody.innerHTML = `<tr><td colspan="8" class="empty-state">No devices yet. Add one using the panel on the left.</td></tr>`;
      return;
    }

    tbody.innerHTML = devices.map(d => `
      <tr id="row-${d.id}">
        <td class="id-cell" title="${d.id}">${d.id.split('-')[0]}...</td>
        <td>${d.imei}</td>
        <td><strong>${d.model}</strong></td>
        <td><span class="status-badge status-${d.status}">${d.status}</span></td>
        <td><span class="grade-badge grade-${d.grade}">${d.grade}</span></td>
        <td>$${parseFloat(d.price).toFixed(2)}</td>
        <td style="color:#64748b">${timeAgo(d.updated_at)}</td>
        <td>
          <div class="action-btns">
            <button class="btn-purple" onclick="quickProcess('${d.id}')">⚡ Grade</button>
            <button class="btn-secondary" onclick="copyId('${d.id}')">Copy ID</button>
          </div>
        </td>
      </tr>
    `).join('');

  } catch (err) {
    toast('Could not reach API. Is the server running?', 'error');
  }
}

// ── Create device ────────────────────────────────────────────────────────────
async function createDevice() {
  const imei  = document.getElementById('imei').value.trim();
  const model = document.getElementById('model').value.trim();
  const price = parseFloat(document.getElementById('price').value);

  if (!imei || !model || isNaN(price)) {
    toast('Please fill in IMEI, Model and Price', 'error');
    return;
  }

  const res = await fetch(`${API}/devices`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ imei, model, price })
  });

  const data = await res.json();

  if (!res.ok) {
    toast(`Error: ${data.error}`, 'error');
    return;
  }

  toast(`✓ Device added: ${data.model}`);
  document.getElementById('imei').value  = '';
  document.getElementById('model').value = '';
  document.getElementById('price').value = '';
  loadDevices();
}

// ── Update status ────────────────────────────────────────────────────────────
async function updateStatus() {
  const id     = document.getElementById('status-id').value.trim();
  const status = document.getElementById('new-status').value;

  if (!id) { toast('Paste a Device ID first', 'error'); return; }

  const res = await fetch(`${API}/devices/${id}/status`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status })
  });

  const data = await res.json();
  if (!res.ok) { toast(`Error: ${data.error}`, 'error'); return; }

  toast(`✓ Status updated to "${status}"`);
  loadDevices();
}

// ── Process / grade a device ─────────────────────────────────────────────────
async function processDevice() {
  const id = document.getElementById('process-id').value.trim();
  if (!id) { toast('Paste a Device ID first', 'error'); return; }
  await runProcessing(id);
}

// Called from the table row button
async function quickProcess(id) {
  await runProcessing(id);
}

async function runProcessing(id) {
  toast('⚡ Goroutine launched — grading in progress...', 'success');

  const res = await fetch(`${API}/devices/${id}/process`, { method: 'POST' });
  const data = await res.json();

  if (!res.ok) { toast(`Error: ${data.error}`, 'error'); return; }

  toast(`✓ ${data.message}`);
  loadDevices();
}

// ── Lookup single device ─────────────────────────────────────────────────────
async function lookupDevice() {
  const id = document.getElementById('lookup-id').value.trim();
  if (!id) { toast('Enter a device ID', 'error'); return; }

  const res = await fetch(`${API}/devices/${id}`);
  const data = await res.json();

  if (!res.ok) { toast('Device not found', 'error'); return; }

  toast(`Found: ${data.model} | Status: ${data.status} | Grade: ${data.grade}`);
}

// ── Helpers ──────────────────────────────────────────────────────────────────
function copyId(id) {
  navigator.clipboard.writeText(id);
  toast('ID copied to clipboard');
}

function timeAgo(dateStr) {
  const diff = Math.floor((Date.now() - new Date(dateStr)) / 1000);
  if (diff < 60)   return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff/60)}m ago`;
  return `${Math.floor(diff/3600)}h ago`;
}

// Auto-refresh every 10 seconds
loadDevices();
setInterval(loadDevices, 10000);