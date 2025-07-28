const API_BASE_URL = 'http://localhost:8000/api/v1';

document.addEventListener("DOMContentLoaded", async () => {
  const token = localStorage.getItem("access_token");
  const container = document.getElementById("books-grid");

  try {
    const res = await fetch(`${API_BASE_URL}/books`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    const result = await res.json();

    if (res.status === 401) {
      alert("Tu sesión ha expirado. Vuelve a iniciar sesión.");
      localStorage.removeItem("access_token");
      window.location.href = "login.html";
      return;
    }

    if (res.ok && result.data && result.data.length > 0) {
      result.data.forEach(book => {
        const canReserve = book.available_exemplars > 0;

        const card = document.createElement("div");
        card.className = "book-card";
        card.innerHTML = `
          <h3>${book.title}</h3>
          <p><strong>Autor:</strong> ${book.author}</p>
          <button onclick="reserve(${book.id})" ${!canReserve ? "disabled" : ""}>
            ${canReserve ? "Reservar" : "No disponible"}
          </button>
        `;
        container.appendChild(card);
      });
    } else {
      container.innerHTML = "<p>No se encontraron libros.</p>";
    }
  } catch (err) {
    container.innerHTML = "<p>Error cargando libros.</p>";
    console.error(err);
  }
});

async function reserve(bookId) {
  const token = localStorage.getItem("access_token");

  if (!token) {
    alert("Debes iniciar sesión.");
    window.location.href = "login.html";
    return;
  }

  try {
    const res = await fetch(`${API_BASE_URL}/loans`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ exemplar_id: bookId }) // usamos bookId como si fuera exemplar_id
    });

    const result = await res.json();

    if (res.status === 401) {
      alert("Tu sesión ha expirado. Vuelve a iniciar sesión.");
      localStorage.removeItem("access_token");
      window.location.href = "login.html";
      return;
    }

    if (res.ok) {
      alert("✅ Libro reservado correctamente.");
      location.reload();
    } else {
      alert(result.error || "No se pudo reservar el libro.");
    }
  } catch (error) {
    console.error(error);
    alert("Error al conectar con el servidor.");
  }
}
