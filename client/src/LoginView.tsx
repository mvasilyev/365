import { useState } from 'react';
import { API } from './api';

export function LoginView() {
    const [username, setUsername] = useState('admin');
    const [status, setStatus] = useState('');

    const handleRegister = async () => {
        try {
            setStatus('Registering...');
            await API.register(username);
            setStatus('Registration successful! You can now login.');
        } catch (e: any) {
            setStatus('Error: ' + e.message);
        }
    };

    const handleLogin = async () => {
        try {
            setStatus('Logging in...');
            await API.login(username);
            setStatus('Login successful!');
            window.location.href = '/'; // Simple redirect
        } catch (e: any) {
            setStatus('Error: ' + e.message);
        }
    };

    return (
        <div style={{ maxWidth: 400, margin: '50px auto', textAlign: 'center' }}>
            <h1>Admin Access</h1>
            <input
                value={username}
                onChange={e => setUsername(e.target.value)}
                placeholder="Username"
                style={{ padding: 10, fontSize: 16, width: '100%', marginBottom: 20, background: '#333', color: '#fff', border: 'none', borderRadius: 8 }}
            />
            <div style={{ display: 'flex', gap: 10 }}>
                <button onClick={handleRegister} style={btnStyle}>Register New Device</button>
                <button onClick={handleLogin} style={{ ...btnStyle, background: '#007bff' }}>Login</button>
            </div>
            {status && <p style={{ marginTop: 20 }}>{status}</p>}
        </div>
    );
}

const btnStyle: React.CSSProperties = {
    flex: 1,
    padding: 15,
    borderRadius: 8,
    border: 'none',
    background: '#444',
    color: '#fff',
    fontSize: 16,
    cursor: 'pointer',
};
