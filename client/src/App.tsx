import { useState, useEffect } from 'react';
import { BrowserRouter, Route, Routes, Link } from 'react-router-dom';
import { GalleryView } from './GalleryView';
import { LoginView } from './LoginView';
import { UploadView } from './UploadView';
import { DetailView } from './DetailView';
import { API } from './api';

type Theme = 'light' | 'dark' | 'system';

function App() {
    const [theme, setTheme] = useState<Theme>(() => {
        return (localStorage.getItem('theme') as Theme) || 'system';
    });
    const [isAuth, setIsAuth] = useState(false);

    useEffect(() => {
        API.checkAuth().then(setIsAuth).catch(() => setIsAuth(false));
    }, []);

    useEffect(() => {
        const root = document.documentElement;
        const systemDark = window.matchMedia('(prefers-color-scheme: dark)');

        const applyTheme = () => {
            if (theme === 'dark' || (theme === 'system' && systemDark.matches)) {
                root.setAttribute('data-theme', 'dark');
            } else {
                root.setAttribute('data-theme', 'light');
            }
        };

        applyTheme();
        localStorage.setItem('theme', theme);

        systemDark.addEventListener('change', applyTheme);
        return () => systemDark.removeEventListener('change', applyTheme);
    }, [theme]);

    const cycleTheme = () => {
        const Map: Record<Theme, Theme> = { light: 'dark', dark: 'system', system: 'light' };
        setTheme(Map[theme]);
    };

    const themeIcon = { light: '☀️', dark: '🌙', system: '⚙️' }[theme];

    return (
        <BrowserRouter>
            <div style={{ display: 'flex', flexDirection: 'column', height: '100dvh', background: 'var(--bg-color)', color: 'var(--text-color)' }}>
                <header style={{ padding: '15px 20px', borderBottom: '1px solid var(--border-color)', background: 'var(--header-bg)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
                        <Link to="/" style={{ textDecoration: 'none', color: 'var(--text-color)', fontWeight: 'bold', fontSize: 20 }}>365</Link>
                        {isAuth && (
                            <Link to="/upload" style={{ textDecoration: 'none', color: 'var(--text-color)', fontSize: 14, border: '1px solid var(--border-color)', padding: '4px 10px', borderRadius: 4 }}>+ Upload</Link>
                        )}
                    </div>

                    <button
                        onClick={cycleTheme}
                        style={{
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            fontSize: 14,
                            padding: 5,
                            display: 'flex',
                            alignItems: 'center',
                            gap: 6,
                            color: 'var(--text-color)'
                        }}
                        title={`Theme: ${theme}`}
                    >
                        <span style={{ opacity: 0.7, marginRight: 4 }}>Theme:</span>
                        <span style={{ fontSize: 18 }}>{themeIcon}</span>
                        <span style={{ textTransform: 'capitalize', fontWeight: 500 }}>{theme}</span>
                    </button>
                </header>

                <main style={{ flex: 1, overflowY: 'auto' }}>
                    <Routes>
                        <Route path="/" element={<GalleryView />} />
                        <Route path="/login" element={<LoginView />} />
                        <Route path="/upload" element={<UploadView />} />
                        <Route path="/day/:date" element={<DetailView />} />
                    </Routes>
                </main>
            </div>
        </BrowserRouter>
    )
}

export default App
