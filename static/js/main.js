import { getToken, sendRequest, setToken } from './api.js';
import { updateDashboard } from './dashboard.js';
import { displayUrls, initSortHandlers } from './table.js';
import { hideResult, showError, showResult, showView } from './ui.js';

async function loadMyUrls() {
  try {
    const urls = (await sendRequest('/my', 'GET')).data;
    const table = document.getElementById('urlsTableBody');
    table.dataset.urls = JSON.stringify(urls);
    displayUrls(urls);
    updateDashboard(urls);
  } catch (error) {
    showError(error.message);
  }
}

// si hay token, ocultar el formulario de login y mostrar el contenido loggeado
if (getToken()) {
  document.getElementById('loginForm').style.display = 'none';
  document.getElementById('loggedInContent').style.display = 'block';
  showView(false);
  loadMyUrls();
}

// login handler
document.getElementById('authForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  try {
    const data = await sendRequest('/auth', 'POST', {
      username: document.getElementById('username').value,
      password: document.getElementById('password').value
    });
    setToken(data.token);
    hideResult();
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('loggedInContent').style.display = 'block';
    showView(false);
    loadMyUrls();
  } catch (error) {
    if (error.message === 'Unauthorized') {
      showError('invalid credentials');
    } else {
      showError(error.message);
    }
  }
});

// handler del form
document.getElementById('urlForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  const url = document.getElementById('longUrl').value;
  const hash = document.getElementById('customHash').value;

  if (hash && !/^[a-zA-Z0-9_-]+$/.test(hash)) {
    showError('custom hash should contain only letters, numbers, underscores and dashes.');
    return;
  }

  try {
    const data = (await sendRequest('/shorten', 'POST', { url, hash })).data;
    const link = document.createElement('a');
    link.href = `${window.location.origin}/${data.hash}`;
    link.textContent = `${window.location.origin}/${data.hash}`;
    link.target = '_blank';

    if (data.existed) {
      const span = document.createElement('span');
      span.textContent = ' (already existed)';
      const div = document.createElement('div');
      div.appendChild(link);
      div.appendChild(span);
      showResult(div, 'warning');
      return;
    }

    showResult(link, 'success');
    loadMyUrls();
  } catch (error) {
    showError(error.message);
  }
});

// handlers de navegación para la navbar
document.getElementById('shortenLink').addEventListener('click', (e) => {
  e.preventDefault();
  hideResult();
  showView(false);
});

document.getElementById('myUrlsLink').addEventListener('click', (e) => {
  e.preventDefault();
  hideResult();
  showView(true);
  loadMyUrls();
});

document.getElementById('logoutLink').addEventListener('click', (e) => {
  e.preventDefault();
  setToken('');
  document.getElementById('loggedInContent').style.display = 'none';
  document.getElementById('dashboard').style.display = 'none';
  document.getElementById('loginForm').style.display = 'block';
  document.getElementById('username').value = '';
  document.getElementById('password').value = '';
  hideResult();
});

// inicializar los handlers de orden en la tabla
document.addEventListener('DOMContentLoaded', initSortHandlers);
