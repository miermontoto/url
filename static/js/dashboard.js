function updateDashboard(urls) {
  const totalHits = urls.reduce((sum, url) => sum + url.hits, 0);
  const totalUrls = urls.length;

  // la cantidad de códigos restantes es la cantidad de códigos posibles (36^3)
  // menos la cantidad de códigos de 3 caracteres ya guardados
  const threeCharUrls = urls.filter(url => url.hash.length === 3).length;
  const availableCodes = Math.pow(36, 3) - threeCharUrls;

  document.getElementById('totalHits').textContent = totalHits.toLocaleString();
  document.getElementById('totalUrls').textContent = totalUrls.toLocaleString();
  document.getElementById('availableCodes').textContent = availableCodes.toLocaleString();

  // solo mostrar en la vista principal
  const isUrlListVisible = document.getElementById('myUrlsList').style.display === 'block';
  document.getElementById('dashboard').style.display = isUrlListVisible ? 'none' : 'block';
}

export { updateDashboard };
