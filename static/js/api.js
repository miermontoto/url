// la autenticación es por cookie httpOnly de sesión (patrón backend-for-frontend):
// el navegador no maneja tokens, solo envía la cookie same-origin en cada petición.

async function sendRequest(url, method, body) {
  const response = await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body)
  });

  if (!response.ok) {
    if (response.status === 401) {
      showLogin();
      throw new Error('session expired - please login again');
    }
    throw new Error(response.statusText);
  }

  return await response.json();
}

// muestra el formulario de login y oculta el contenido privado
function showLogin() {
  document.getElementById('loggedInContent').style.display = 'none';
  document.getElementById('dashboard').style.display = 'none';
  document.getElementById('loginForm').style.display = 'block';
}

async function deleteUrl(hash) {
  return await sendRequest(`/my/${hash}`, 'DELETE');
}

export { deleteUrl, sendRequest, showLogin };
