const API_BASE_URL = 'http://localhost:8000/api/v1'; // Asegúrate que sea tu puerto real

// Renderiza los libros prestados
function renderMyBooks(data) {
  const tbody = document.getElementById('mybooks-tbody');
  tbody.innerHTML = ''; // Limpiar tabla

  if (!data || data.length === 0) {
    tbody.innerHTML = '<tr><td colspan="4">No tienes préstamos registrados.</td></tr>';
    return;
  }

  const today = new Date();
  
  data.forEach(res => {
    const dueDate = new Date(res.due_date);
    if (dueDate < today) return; // No mostrar si la fecha ya pasó

    const row = document.createElement('tr');
    row.className = calculateBookStatus(res.due_date);

    row.innerHTML = `
      <td>${res.exemplar_code || ''}</td>
      <td>${res.book_title || ''}</td>
      <td>${res.loan_date?.split("T")[0] || ''}</td>
      <td>${res.due_date?.split("T")[0] || ''}</td>
    `;
    tbody.appendChild(row);
  });
}

// Determina el estado del libro en función de la fecha de devolución
function calculateBookStatus(returnDate) {
  if (!returnDate) return 'normal';

  const today = new Date();
  const date = new Date(returnDate);
  const diff = Math.ceil((date - today) / (1000 * 60 * 60 * 24));

  if (diff < 0) return 'overdue';
  if (diff <= 3) return 'due-soon';
  return 'normal';
}

// Obtiene libros prestados desde el backend
async function loadMyBooks() {
  const token = localStorage.getItem("access_token");

  if (!token) {
    window.location.href = "login.html";
    return;
  }

  try {
    const response = await fetch(`${API_BASE_URL}/loans/my`, {
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
    renderMyBooks(result.data);
  } catch (error) {
    console.error("Error al obtener los préstamos:", error);
    document.getElementById('error-message').textContent =
      "No se pudieron cargar tus préstamos. Verifica tu sesión o contacta al administrador.";
  }
}

// Cierra sesión
function logout() {
  if (confirm('¿Salir de la sesión?')) {
    localStorage.removeItem("access_token");
    window.location.href = "login.html";
  }
}

// Inicia la carga al abrir la página
document.addEventListener('DOMContentLoaded', loadMyBooks);
