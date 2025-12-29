import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { API, Photo } from './api';
import { Map, Marker } from 'pigeon-maps';

export function DetailView() {
    const { date } = useParams<{ date: string }>();
    const [photo, setPhoto] = useState<Photo | null>(null);
    const [exif, setExif] = useState<Record<string, string>>({});

    const [prevDay, setPrevDay] = useState<string | null>(null);
    const [nextDay, setNextDay] = useState<string | null>(null);

    useEffect(() => {
        if (!date) return;
        API.getPhotos().then(photos => {
            const idx = photos.findIndex(p => p.Day === date);
            if (idx >= 0) {
                setPhoto(photos[idx]);
                try {
                    setExif(JSON.parse(photos[idx].ExifData || '{}'));
                } catch { setExif({}) }

                // List is DESC (Newest first)
                // Next (Newer) is idx - 1
                // Prev (Older) is idx + 1
                if (idx > 0) setNextDay(photos[idx - 1].Day);
                else setNextDay(null);

                if (idx < photos.length - 1) setPrevDay(photos[idx + 1].Day);
                else setPrevDay(null);
            }
        });
    }, [date]);

    if (!photo) return <div style={{ padding: 20 }}>Loading or not found... <Link to="/">Back</Link></div>;

    // Filter interesting EXIF
    const exifKeys = [
        { label: 'Camera', key: 'Model' },
        { label: 'Lens', key: 'LensModel' },
        { label: 'F-Stop', key: 'FNumber' },
        { label: 'ISO', key: 'ISOSpeedRatings' },
        { label: 'Shutter', key: 'ExposureTime' },
        { label: 'Date', key: 'DateTimeOriginal' },
    ];

    const hasCoords = photo.Lat !== 0 || photo.Lon !== 0;

    return (
        <div style={{ display: 'flex', flexDirection: 'column', height: '100%', overflowY: 'auto', background: 'var(--bg-color)', color: 'var(--text-color)' }}>
            <div style={{ padding: '10px 20px', borderBottom: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Link to="/" style={{ color: 'var(--text-muted)', textDecoration: 'none', width: 60 }}>&uarr; Gallery</Link>

                <div style={{ display: 'flex', gap: 20, alignItems: 'center' }}>
                    {prevDay ? (
                        <Link to={`/day/${prevDay}`} style={{ color: 'var(--text-color)', textDecoration: 'none' }}>&larr; Prev</Link>
                    ) : (
                        <span style={{ color: 'var(--text-muted)' }}>&larr; Prev</span>
                    )}

                    <span style={{ fontWeight: 'bold' }}>{photo.Day}</span>

                    {nextDay ? (
                        <Link to={`/day/${nextDay}`} style={{ color: 'var(--text-color)', textDecoration: 'none' }}>Next &rarr;</Link>
                    ) : (
                        <span style={{ color: 'var(--text-muted)' }}>Next &rarr;</span>
                    )}
                </div>

                <div style={{ width: 60 }}></div>
            </div>

            {/* Main Image */}
            <div style={{ padding: 20, display: 'flex', justifyContent: 'center', background: 'var(--card-bg)' }}>
                <img
                    src={photo.Filepath}
                    alt={photo.Day}
                    style={{ maxHeight: '60vh', maxWidth: '100%', objectFit: 'contain' }}
                />
            </div>

            <div style={{ padding: 20, display: 'flex', flexWrap: 'wrap', gap: 20 }}>
                {/* Notes */}
                <div style={{ flex: '1 1 300px' }}>
                    <h3 style={{ borderBottom: '1px solid var(--border-color)', paddingBottom: 5, marginTop: 0 }}>Notes</h3>
                    <div style={{ whiteSpace: 'pre-wrap', lineHeight: '1.5', color: 'var(--text-color)' }}>
                        {photo.Notes || "No notes."}
                    </div>
                </div>

                {/* EXIF */}
                <div style={{ flex: '1 1 200px' }}>
                    <h3 style={{ borderBottom: '1px solid var(--border-color)', paddingBottom: 5, marginTop: 0 }}>Details</h3>
                    <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '8px 15px', fontSize: 14 }}>
                        {exifKeys.map(({ label, key }) => {
                            const val = exif[key];
                            if (!val) return null;
                            return (
                                <div key={key} style={{ display: 'contents' }}>
                                    <div style={{ color: 'var(--text-muted)' }}>{label}</div>
                                    <div style={{ color: 'var(--text-color)' }}>{cleanExifVal(val)}</div>
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Map */}
                {hasCoords && (
                    <div style={{ flex: '1 1 300px', height: 300, borderRadius: 8, overflow: 'hidden', border: '1px solid var(--border-color)' }}>
                        <Map height={300} defaultCenter={[photo.Lat, photo.Lon]} defaultZoom={13}>
                            <Marker width={50} anchor={[photo.Lat, photo.Lon]} />
                        </Map>
                    </div>
                )}
            </div>
        </div>
    );
}

// Clean up some raw EXIF values (like arrays/quotes)
function cleanExifVal(val: string): string {
    if (val.startsWith('"') && val.endsWith('"')) {
        return val.slice(1, -1);
    }
    return val;
}
