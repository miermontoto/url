function showView(isUrlList) {
    document.getElementById('shortenerForm').style.display = isUrlList ? 'none' : 'block';
    document.getElementById('myUrlsList').style.display = isUrlList ? 'block' : 'none';
    document.getElementById('dashboard').style.display = isUrlList ? 'none' : 'block';
    document.getElementById('shortenLink').classList.toggle('active', !isUrlList);
    document.getElementById('myUrlsLink').classList.toggle('active', isUrlList);
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

export { showView, showError, showResult, hideResult };
