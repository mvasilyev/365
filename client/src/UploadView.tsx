import { useState } from 'react';
import { API } from './api';

export function UploadView() {
    const [file, setFile] = useState<File | null>(null);
    const [day, setDay] = useState(new Date().toISOString().split('T')[0]);
    const [notes, setNotes] = useState('');
    const [status, setStatus] = useState('');

    const handleUpload = async () => {
        if (!file) return;
        try {
            setStatus('Uploading...');
            await API.uploadPhoto(file, day, notes);
            setStatus('Uploaded!');
            setFile(null);
            setNotes('');
        } catch (e: any) {
            setStatus('Error: ' + e.message);
        }
    };

    return (
        <div style={{ padding: 20, maxWidth: 600, margin: '0 auto' }}>
            <h2>Upload Photo</h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 15 }}>
                <input type="date" value={day} onChange={e => setDay(e.target.value)} style={inputStyle} />
                <input type="file" accept="image/*" onChange={e => setFile(e.target.files?.[0] || null)} style={inputStyle} />
                <textarea
                    value={notes}
                    onChange={e => setNotes(e.target.value)}
                    placeholder="Notes..."
                    rows={4}
                    style={inputStyle}
                />
                <button onClick={handleUpload} style={btnStyle}>Upload</button>
                {status && <p>{status}</p>}
            </div>
        </div>
    );
}

const inputStyle = {
    padding: 12,
    background: '#222',
    color: '#fff',
    border: '1px solid #444',
    borderRadius: 6,
    fontSize: 16,
};

const btnStyle = {
    padding: 15,
    background: '#eee',
    color: '#000',
    border: 'none',
    borderRadius: 6,
    fontSize: 16,
    cursor: 'pointer',
    fontWeight: 'bold',
};
