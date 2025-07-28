// Datos de ejemplo para las recomendaciones
const recommendationsData = [
  {
    id: 1,
    title: "Mastering Java EE",
    author: "Marifé Dickinson",
    image: "200 x 200"
  },
  {
    id: 2,
    title: "Thinking in Java",
    author: "Winchester McFry",
    image: "200 x 200"
  },
  {
    id: 3,
    title: "Wildfly in action",
    author: "Arun Gupta",
    image: "200 x 200"
  }
];

// Datos del libro actual (ejemplo)
const currentBook = {
  id: 10,
  title: "Head First Design Patterns",
  author: "Freeman & Freeman",
  isbn: "1234567890",
  pages: "180",
  copies: "10",
  image: "200 x 400"
};

function loadBookDetails() {
  // En una aplicación real, obtendrías estos datos de la URL o API
  const urlParams = new URLSearchParams(window.location.search);
  const bookId = urlParams.get('id');
  
  // Por ahora, usamos datos de ejemplo
  document.getElementById('book-title').textContent = currentBook.title;
  document.getElementById('book-author').textContent = currentBook.author;
  document.getElementById('book-isbn').textContent = currentBook.isbn;
  document.getElementById('book-pages').textContent = currentBook.pages;
  document.getElementById('book-copies').textContent = currentBook.copies;
}

function renderRecommendations() {
  const recommendationsGrid = document.getElementById('recommendations-grid');
  
  recommendationsGrid.innerHTML = recommendationsData.map(book => `
    <div class="recommendation-card">
      <div class="recommendation-image">
        ${book.image}
      </div>
      <div class="recommendation-info">
        <div class="recommendation-title">${book.title}</div>
        <div class="recommendation-author">${book.author}</div>
        <div class="recommendation-actions">
          <a href="#" class="btn-detail" onclick="viewDetail(${book.id})">Ver detalle →</a>
        </div>
      </div>
    </div>
  `).join('');
}

function reserveBook() {
  // Implementar lógica de reserva
  if (confirm('¿Deseas reservar este libro?')) {
    alert('Libro reservado exitosamente. Te notificaremos cuando esté disponible.');
    // Aquí puedes agregar la lógica para enviar la reserva al servidor
  }
}

function viewDetail(bookId) {
  // Navegar al detalle de otro libro
  window.location.href = `book-detail.html?id=${bookId}`;
}

function logout() {
  if (confirm('¿Estás seguro de que quieres salir?')) {
    // Implementar lógica de logout
    window.location.href = 'index.html';
  }
}

// Inicializar la página
document.addEventListener('DOMContentLoaded', function() {
  loadBookDetails();
  renderRecommendations();
});