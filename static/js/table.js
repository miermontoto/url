let currentSort = { column: 'date', ascending: false };

function formatDate(date) {
  return new Date(date).toLocaleString('sv', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

function displayUrls(urls) {
  const table = document.getElementById('urlsTableBody');
  table.innerHTML = '';

  // ordenar las urls
  urls.sort((a, b) => {
    let aVal = currentSort.column === 'date' ? new Date(a.created).getTime() : a.hits;
    let bVal = currentSort.column === 'date' ? new Date(b.created).getTime() : b.hits;
    return (currentSort.ascending ? 1 : -1) * (aVal - bVal);
  });

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
    dateCell.textContent = formatDate(url.created);
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

  // actualizar los indicadores de orden
  document.querySelectorAll('th.sortable').forEach(th => {
    const column = th.dataset.sort;
    if (column === currentSort.column) {
      th.textContent = `${column} ${currentSort.ascending ? '↑' : '↓'}`;
    } else {
      th.textContent = `${column} ↕`;
    }
  });
}

function initSortHandlers() {
  document.querySelectorAll('th.sortable').forEach(th => {
    th.addEventListener('click', () => {
      const column = th.dataset.sort;
      if (currentSort.column === column) {
        currentSort.ascending = !currentSort.ascending;
      } else {
        currentSort.column = column;
        currentSort.ascending = true;
      }
      const urls = JSON.parse(document.getElementById('urlsTableBody').dataset.urls);
      displayUrls(urls);
    });
  });
}

export { displayUrls, initSortHandlers };
