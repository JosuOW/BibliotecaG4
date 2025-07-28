const API_BASE_URL = 'http://localhost:8000/api/v1';

// Renderiza el historial de préstamos del usuario
function renderHistory(data) {
  const tbody = document.querySelector('#history-table tbody');
  tbody.innerHTML = '';

  if (!data || data.length === 0) {
    tbody.innerHTML = '<tr><td colspan="5">No tienes historial de préstamos.</td></tr>';
    return;
  }

  data.forEach(res => {
    const row = document.createElement('tr');
    row.className = calculateBookStatus(res.due_date);

    row.innerHTML = `
      <td>${res.book_title || ''}</td>
      <td>${res.exemplar_code || ''}</td>
      <td>${res.loan_date?.split("T")[0] || ''}</td>
      <td>${res.due_date?.split("T")[0] || ''}</td>
      <td>${res.is_overdue ? `Sí (${res.overdue_days} días)` : 'No'}</td>
    `;
    tbody.appendChild(row);
  });
}

function calculateBookStatus(dueDate) {
  if (!dueDate) return 'normal';

  const today = new Date();
  const date = new Date(dueDate);
  const diff = Math.ceil((date - today) / (1000 * 60 * 60 * 24));

  if (diff < 0) return 'overdue';
  if (diff <= 3) return 'due-soon';
  return 'normal';
}

async function loadHistory() {
  const token = localStorage.getItem("access_token");
  if (!token) {
    window.location.href = "login.html";
    return;
  }

  try {
    const response = await fetch(`${API_BASE_URL}/loans/history`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    if (response.status === 401) {
      localStorage.removeItem("access_token");
      alert("Tu sesión ha expirado. Por favor, vuelve a iniciar sesión.");
      window.location.href = "login.html";
      return;
    }

    if (!response.ok) {
      const text = await response.text();
      throw new Error(`Error ${response.status}: ${text}`);
    }

    const result = await response.json();
    renderHistory(result.data);
  } catch (error) {
    console.error("Error al obtener el historial:", error);
    document.getElementById('error-message').textContent =
      "No se pudo cargar el historial. Intenta más tarde o contacta al administrador.";
  }
}

document.addEventListener('DOMContentLoaded', loadHistory);
