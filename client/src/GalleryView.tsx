import { useEffect, useState } from 'react';
import { API, Photo } from './api';
import { Link } from 'react-router-dom';

export function GalleryView() {
    const [photos, setPhotos] = useState<Photo[]>([]);
    const [currentMonthIdx, setCurrentMonthIdx] = useState(0);

    useEffect(() => {
        API.getPhotos().then(setPhotos).catch(console.error);
    }, []);

    // Group by Month: "YYYY-MM" -> Photo[]
    const months = (photos || []).reduce((acc, p) => {
        const key = p.Day.substring(0, 7); // 2023-12
        if (!acc[key]) acc[key] = [];
        acc[key].push(p);
        return acc;
    }, {} as Record<string, Photo[]>);

    const sortedMonths = Object.keys(months).sort().reverse();
    const currentMonth = sortedMonths[currentMonthIdx];

    const hasPrev = currentMonthIdx < sortedMonths.length - 1; // Older
    const hasNext = currentMonthIdx > 0; // Newer

    const handlePrev = () => { if (hasPrev) setCurrentMonthIdx(currentMonthIdx + 1); };
    const handleNext = () => { if (hasNext) setCurrentMonthIdx(currentMonthIdx - 1); };

    if (!photos.length) return <p style={{ textAlign: 'center', opacity: 0.5, marginTop: 50 }}>No photos yet.</p>;
    if (!currentMonth) return null;

    return (
        <div style={{ padding: '20px', maxWidth: 1200, margin: '0 auto' }}>

            {/* Navigation Header */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
                <button
                    onClick={handlePrev}
                    disabled={!hasPrev}
                    style={{ ...navBtnStyle, opacity: hasPrev ? 1 : 0.3 }}
                >
                    &larr; Prev
                </button>

                <h2 style={{ margin: 0, color: 'var(--text-color)' }}>
                    {new Date(currentMonth + '-01').toLocaleString('default', { month: 'long', year: 'numeric' })}
                </h2>

                <button
                    onClick={handleNext}
                    disabled={!hasNext}
                    style={{ ...navBtnStyle, opacity: hasNext ? 1 : 0.3 }}
                >
                    Next &rarr;
                </button>
            </div>

            <MonthSection month={currentMonth} photos={months[currentMonth]} />
        </div>
    );
}

const navBtnStyle: React.CSSProperties = {
    background: 'var(--btn-bg)',
    border: '1px solid var(--border-color)',
    color: 'var(--btn-text)',
    padding: '8px 16px',
    borderRadius: 4,
    cursor: 'pointer',
    fontSize: 14,
};

function MonthSection({ month, photos }: { month: string, photos: Photo[] }) {
    // month is "YYYY-MM"
    const [year, m] = month.split('-').map(Number);
    const date = new Date(year, m - 1, 1);
    const monthName = date.toLocaleString('default', { month: 'long', year: 'numeric' });

    // Calculate padding for Monday start
    // getDay(): Sun=0, Mon=1, ..., Sat=6
    // We want Mon=0, ..., Sun=6
    // (day + 6) % 7
    const startDay = (date.getDay() + 6) % 7;

    // Create placeholders
    const placeholders = Array.from({ length: startDay });

    // Create a map of day -> photo for O(1) lookup
    const photoMap = new Map(photos.map(p => [p.Day, p]));

    // Days in month
    const daysInMonth = new Date(year, m, 0).getDate();
    const days = Array.from({ length: daysInMonth }, (_, i) => {
        const d = i + 1;
        const dayStr = `${month}-${String(d).padStart(2, '0')}`;
        return { dayStr, photo: photoMap.get(dayStr), dayNum: d };
    });

    return (
        <div style={{ marginBottom: 40 }}>
            <h2 style={{ color: 'var(--text-muted)', borderBottom: '1px solid var(--border-color)', paddingBottom: 5, marginBottom: 10 }}>{monthName}</h2>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 2 }}>
                {/* Headers */}
                {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map(d => (
                    <div key={d} style={{ textAlign: 'center', fontSize: 10, color: 'var(--text-muted)', paddingBottom: 5 }}>{d}</div>
                ))}

                {/* Placeholders */}
                {placeholders.map((_, i) => <div key={`ph-${i}`} />)}

                {/* Days */}
                {days.map(d => (
                    <div key={d.dayStr} style={{ aspectRatio: '1', position: 'relative', background: 'var(--card-bg)', borderRadius: 4, overflow: 'hidden' }}>
                        {d.photo ? (
                            <Link to={`/day/${d.photo.Day}`} style={{ display: 'block', width: '100%', height: '100%' }}>
                                <img
                                    src={d.photo.ThumbnailPath || d.photo.Filepath}
                                    alt={d.photo.Day}
                                    style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                                />
                            </Link>
                        ) : (
                            <div style={{ opacity: 0.1, display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>

                            </div>
                        )}
                        <div style={{
                            position: 'absolute', top: 2, right: 2,
                            fontSize: 10, fontWeight: 'bold',
                            color: d.photo ? '#fff' : 'var(--text-muted)',
                            textShadow: d.photo ? '0 1px 2px rgba(0,0,0,0.8)' : 'none'
                        }}>
                            {d.dayNum}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
