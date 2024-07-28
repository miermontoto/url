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
	const customHash = document.getElementById('customHash').value;

	if (customHash && !/^[a-zA-Z0-9_-]+$/.test(customHash)) {
		showError('custom hash should contain only letters, numbers, underscores and dashes.');
		return;
	}

	try {
		const response = await fetch('/shorten', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'Authorization': `Bearer ${token}`
			},
			body: JSON.stringify({
				url: longUrl,
				customHash: customHash || undefined
			})
		});

		if (!response.ok) {
			showError(response.statusText);
			return;
		}

		const data = (await response.json()).data;

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
			return
		}

		showResult(link, 'success');
	} catch (error) {
		showError(error.message);
	}
});

function showError(message) {
	const messageElement = document.createElement('span');
	messageElement.textContent = message;
	showResult(messageElement, 'danger');
}

function showResult(message, type) {
	const resultElement = document.getElementById('result');
	resultElement.innerHTML = '';
	resultElement.className = `alert alert-${type}`;
	resultElement.style.display = 'block';

	resultElement.appendChild(message);
}
