let token = '';

document.getElementById('authForm').addEventListener('submit', async (e) => {
	e.preventDefault();
	const username = document.getElementById('username').value;
	const password = document.getElementById('password').value;
	try {
		const response = await fetch('/auth', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username, password })
		});

		if (!response.ok) {
			showError(response.statusText);
		}

		const data = await response.json();
		token = data.token;
		document.getElementById('loginForm').style.display = 'none';
		document.getElementById('shortenerForm').style.display = 'block';
	} catch (error) {
		showResult('Error: ' + error.message, 'danger');
	}
});

document.getElementById('urlForm').addEventListener('submit', async (e) => {
	e.preventDefault();
	const longUrl = document.getElementById('longUrl').value;
	try {
		const response = await fetch('/shorten', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'Authorization': token
			},
			body: JSON.stringify({ url: longUrl })
		});

		if (!response.ok) {
			showError(response.statusText);
			return;
		}

		const data = (await response.json()).data;
		if (data.existed) {
			showResult(`${window.location.origin}/${data.hash}\n(already existed)`, 'primary');
			return
		}
		showResult(`successfully generated URL: ${window.location.origin}/${data.hash}`, 'success');
	} catch (error) {
		showError(error.message);
	}
});

function showError(message) {
	showResult(message, 'danger');
}

function showResult(message, type) {
	const resultElement = document.getElementById('result');
	resultElement.textContent = message;
	resultElement.className = `alert alert-${type}`;
	resultElement.style.display = 'block';
}
