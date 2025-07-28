
// auth-utils.js - Utilidades de autenticación para usar en todas las páginas

const KEYCLOAK_CONFIG = {
    KEYCLOAK_URL: 'http://localhost:8081',
    KEYCLOAK_REALM: 'biblioteca',
    KEYCLOAK_CLIENT_ID: 'biblioteca-frontend-dev',
    API_BASE_URL: 'http://localhost:8000/api/v1'
};

// Función para guardar datos de autenticación
function saveAuthData(tokenResponse) {
    localStorage.setItem('access_token', tokenResponse.access_token);
    localStorage.setItem('refresh_token', tokenResponse.refresh_token);
    localStorage.setItem('token_expires_in', tokenResponse.expires_in);
    localStorage.setItem('token_timestamp', Date.now().toString());

    const userInfo = parseUserInfo(tokenResponse.access_token);
    if (userInfo) {
        localStorage.setItem('user_info', JSON.stringify(userInfo));
    }
}

// Función para decodificar el token y extraer info del usuario
function parseUserInfo(accessToken) {
    try {
        const payload = accessToken.split('.')[1];
        const decoded = JSON.parse(atob(payload));
        return {
            email: decoded.email,
            name: decoded.name || decoded.given_name || '',
            lastName: decoded.family_name || '',
            username: decoded.preferred_username,
            roles: decoded.realm_access?.roles || [],
            clientRoles: decoded.resource_access?.[KEYCLOAK_CONFIG.KEYCLOAK_CLIENT_ID]?.roles || []
        };
    } catch (error) {
        console.error('Error parsing token:', error);
        return null;
    }
}

function isTokenExpired() {
    const timestamp = localStorage.getItem('token_timestamp');
    const expiresIn = localStorage.getItem('token_expires_in');
    if (!timestamp || !expiresIn) return true;
    const age = (Date.now() - parseInt(timestamp)) / 1000;
    return age >= parseInt(expiresIn) - 30;
}

async function refreshToken() {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) return false;

    const url = `${KEYCLOAK_CONFIG.KEYCLOAK_URL}/realms/${KEYCLOAK_CONFIG.KEYCLOAK_REALM}/protocol/openid-connect/token`;
    const body = new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
        client_id: KEYCLOAK_CONFIG.KEYCLOAK_CLIENT_ID
    });

    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body
        });
        if (response.ok) {
            const data = await response.json();
            saveAuthData(data);
            return true;
        }
    } catch (e) {
        console.error('Error refreshing token', e);
    }
    return false;
}

async function getValidToken() {
    let token = localStorage.getItem('access_token');
    if (!token) return null;
    if (isTokenExpired()) {
        const refreshed = await refreshToken();
        if (!refreshed) {
            logout();
            return null;
        }
        token = localStorage.getItem('access_token');
    }
    return token;
}

function logout() {
    localStorage.clear();
    window.location.href = 'login.html';
}

async function authenticatedFetch(url, options = {}) {
    const token = await getValidToken();
    if (!token) {
        logout();
        return null;
    }

    const defaultOptions = {
        headers: {
            'Authorization': 'Bearer ' + token,
            'Content-Type': 'application/json',
            ...options.headers
        }
    };

    const response = await fetch(url, { ...options, ...defaultOptions });

    if (response.status === 401) {
        const refreshed = await refreshToken();
        if (refreshed) {
            return authenticatedFetch(url, options);
        } else {
            logout();
        }
    }

    return response;
}

function getUserInfo() {
    const user = localStorage.getItem('user_info');
    return user ? JSON.parse(user) : null;
}

function hasRole(role) {
    const user = getUserInfo();
    return user && (user.roles.includes(role) || user.clientRoles.includes(role));
}

async function isAuthenticated() {
    const token = await getValidToken();
    return !!token;
}

async function requireAuth() {
    const authed = await isAuthenticated();
    if (!authed) {
        window.location.href = 'login.html';
        return false;
    }
    return true;
}

async function requireAdmin() {
    const authed = await requireAuth();
    if (!authed) return false;
    if (!hasRole('ADMIN')) {
        alert('No tienes permisos para esta sección');
        window.location.href = 'index.html';
        return false;
    }
    return true;
}

window.AUTH = {
    getValidToken,
    logout,
    getUserInfo,
    hasRole,
    isAuthenticated,
    requireAuth,
    requireAdmin,
    authenticatedFetch
};

window.API = {}; // Puedes poblarlo luego con tus funciones de backend

document.addEventListener('DOMContentLoaded', async () => {
    if (!window.location.pathname.includes('login.html')) {
        const authed = await isAuthenticated();
        if (!authed) {
            window.location.href = 'login.html';
        }
    }
});
