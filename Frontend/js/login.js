const KEYCLOAK_CONFIG = {
  url: 'http://localhost:8081',
  realm: 'biblioteca',
  clientId: 'biblioteca-frontend-dev',
  clientSecret: 'JCZWqBcae0LOg6SoG4neztHSIsNvwgkT' // <-- usa tu secret real
};

const tokenUrl = `${KEYCLOAK_CONFIG.url}/realms/${KEYCLOAK_CONFIG.realm}/protocol/openid-connect/token`;

document.addEventListener('DOMContentLoaded', () => {
  const loginForm = document.getElementById('login-form');
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;

    if (!email || !password) {
      alert('Completa todos los campos.');
      return;
    }

    const formData = new URLSearchParams();
    formData.append('client_id', KEYCLOAK_CONFIG.clientId);
    formData.append('client_secret', KEYCLOAK_CONFIG.clientSecret); // <-- obligatorio si "Client authentication" está activado
    formData.append('grant_type', 'password');
    formData.append('username', email);
    formData.append('password', password);

    try {
      const response = await fetch(tokenUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData
      });

      const data = await response.json();

      if (!response.ok) {
        console.error('Respuesta del servidor:', data);
        alert(data.error_description || 'Credenciales inválidas.');
        return;
      }

      // Guardar tokens y datos del usuario
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('token_expires_in', data.expires_in);
      localStorage.setItem('token_timestamp', Date.now().toString());

      const payload = JSON.parse(atob(data.access_token.split('.')[1]));
      localStorage.setItem('user_info', JSON.stringify(payload));

      // Redirigir al index
      window.location.href = 'index.html';
    } catch (err) {
      console.error('Error de conexión:', err);
      alert('Error de conexión al autenticar.');
    }
  });
});
