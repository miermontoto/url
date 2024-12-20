let token = localStorage.getItem('token') || '';

function getToken() {
  return token;
}

function setToken(newToken) {
  token = newToken;
  if (newToken) {
    localStorage.setItem('token', newToken);
  } else {
    localStorage.removeItem('token');
  }
}

async function sendRequest(url, method, body) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    method,
    headers,
    body: JSON.stringify(body)
  });

  if (!response.ok) {
    if (response.status === 401) {
      setToken('');
      document.getElementById('loggedInContent').style.display = 'none';
      document.getElementById('dashboard').style.display = 'none';
      document.getElementById('loginForm').style.display = 'block';
      document.getElementById('username').value = '';
      document.getElementById('password').value = '';
      throw new Error('session expired - please login again');
    }
    throw new Error(response.statusText);
  }

  const data = await response.json();
  return data;
}

async function deleteUrl(hash) {
  return await sendRequest(`/my/${hash}`, 'DELETE');
}

export { deleteUrl, getToken, sendRequest, setToken };
