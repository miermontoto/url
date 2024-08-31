let token = '';

document.getElementById('authForm').addEventListener('submit', async (e) => {
	e.preventDefault();
	const username = document.getElementById('username').value;
	const password = document.getElementById('password').value;
	try {
		const data = await sendRequest('/auth', 'POST', { username, password });
		token = data.token;
		document.getElementById('loginForm').style.display = 'none';
		document.getElementById('loggedInContent').style.display = 'block';
	} catch (error) {
		showResult('Error: ' + error.message, 'danger');
	}
});

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
		showError(response.statusText);
	}

	return response.json();
}

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
			return
		}

		showResult(link, 'success');
	} catch (error) {
		showError(error.message);
	}
});

document.getElementById('shortenLink').addEventListener('click', (e) => {
	e.preventDefault();
	hideResult();
	document.getElementById('shortenerForm').style.display = 'block';
	document.getElementById('myUrlsList').style.display = 'none';
	document.getElementById('shortenLink').classList.add('active');
	document.getElementById('myUrlsLink').classList.remove('active');
});

document.getElementById('myUrlsLink').addEventListener('click', (e) => {
	e.preventDefault();
	hideResult();
	document.getElementById('shortenerForm').style.display = 'none';
	document.getElementById('myUrlsList').style.display = 'block';
	document.getElementById('shortenLink').classList.remove('active');
	document.getElementById('myUrlsLink').classList.add('active');
	loadMyUrls();
});

async function loadMyUrls() {
	try {
		const urls = (await sendRequest('/my', 'GET')).data;
		const table = document.getElementById('urlsTableBody');

		// wipe data
		table.innerHTML = '';

		urls.forEach(url => {
			const row = document.createElement('tr');

			const hashCell = document.createElement('td');
			const link = document.createElement('a');
			link.href = `${window.location.origin}/${url.hash}`;
			link.textContent = url.hash;
			link.target = '_blank';
			hashCell.appendChild(link);
			row.appendChild(hashCell);

			const targetCell = document.createElement('td');
			const targetLink = document.createElement('a');
			targetLink.href = url.target;
			targetLink.textContent = url.target;
			targetLink.target = '_blank';
			targetCell.appendChild(targetLink);
			row.appendChild(targetCell);

			const hitsCell = document.createElement('td');
			hitsCell.textContent = url.hits;
			row.appendChild(hitsCell);

			const dateCell = document.createElement('td');
			dateCell.textContent = new Date(url.created).toLocaleString();
			row.appendChild(dateCell);

			const actionsDiv = document.createElement('div');
			actionsDiv.className = 'btn-group';

			const deleteButton = document.createElement('button');
			deleteButton.className = 'btn btn-danger';
			deleteButton.textContent = '🗑';
			deleteButton.onclick = async () => {
				try {
					await deleteUrl(url.hash);
					loadMyUrls();
				} catch (error) {
					showError(error.message);
				}
			};
			actionsDiv.appendChild(deleteButton);

			const actionsRow = document.createElement('td');
			actionsRow.appendChild(actionsDiv);
			row.appendChild(actionsRow);

			table.appendChild(row);
		});
	} catch (error) {
		showError(error.message);
	}
}

async function deleteUrl(hash) {
	return await sendRequest(`/my/${hash}`, 'DELETE');
}


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

function hideResult() {
	const resultElement = document.getElementById('result');
	resultElement.style.display = 'none';
}
