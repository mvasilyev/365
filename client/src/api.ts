
export interface Photo {
    Day: string;
    ID: string;
    Filepath: string;
    ThumbnailPath: string;
    Lat: number;
    Lon: number;
    Notes: string;
    ExifData: string;
}

export const API = {
    async getPhotos(): Promise<Photo[]> {
        const res = await fetch('/api/photos');
        if (!res.ok) throw new Error('Failed to fetch photos');
        return res.json();
    },

    async uploadPhoto(file: File, day: string, notes: string): Promise<void> {
        const formData = new FormData();
        formData.append('photo', file);
        formData.append('day', day);
        formData.append('notes', notes);
        const res = await fetch('/api/photos', {
            method: 'POST',
            body: formData,
        });
        if (!res.ok) throw new Error(await res.text());
    },

    // Auth methods will be added here (WebAuthn is complex, might use a library or raw API)
    async checkAuth(): Promise<boolean> {
        const res = await fetch('/api/auth/status');
        return res.ok;
    },

    async register(username: string) {
        // 1. Get options
        const res = await fetch(`/api/auth/register/begin/${username}`, { method: 'POST' });
        if (!res.ok) throw new Error(await res.text());
        const options = await res.json();

        // 2. Decode options
        options.publicKey.challenge = base64URLToBuffer(options.publicKey.challenge);
        options.publicKey.user.id = base64URLToBuffer(options.publicKey.user.id);

        // 3. Create credentials
        const credential = await navigator.credentials.create({ publicKey: options.publicKey });
        if (!credential) throw new Error("Credential creation failed");

        // 4. Send response
        const cred = credential as PublicKeyCredential;
        const response = {
            id: cred.id,
            rawId: bufferToBase64URL(cred.rawId),
            type: cred.type,
            response: {
                attestationObject: bufferToBase64URL((cred.response as AuthenticatorAttestationResponse).attestationObject),
                clientDataJSON: bufferToBase64URL(cred.response.clientDataJSON),
            },
        };

        const finishRes = await fetch(`/api/auth/register/finish/${username}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
        });
        if (!finishRes.ok) throw new Error(await finishRes.text());
    },

    async login(username: string) {
        // 1. Get options
        const res = await fetch(`/api/auth/login/begin/${username}`, { method: 'POST' });
        if (!res.ok) throw new Error(await res.text());
        const options = await res.json();

        // 2. Decode options
        options.publicKey.challenge = base64URLToBuffer(options.publicKey.challenge);
        options.publicKey.allowCredentials.forEach((c: any) => {
            c.id = base64URLToBuffer(c.id);
        });

        // 3. Get assertion
        const assertion = await navigator.credentials.get({ publicKey: options.publicKey });
        if (!assertion) throw new Error("Assertion failed");

        // 4. Send response
        const cred = assertion as PublicKeyCredential;
        const response = {
            id: cred.id,
            rawId: bufferToBase64URL(cred.rawId),
            type: cred.type,
            response: {
                authenticatorData: bufferToBase64URL((cred.response as AuthenticatorAssertionResponse).authenticatorData),
                clientDataJSON: bufferToBase64URL(cred.response.clientDataJSON),
                signature: bufferToBase64URL((cred.response as AuthenticatorAssertionResponse).signature),
                userHandle: (cred.response as AuthenticatorAssertionResponse).userHandle ? bufferToBase64URL((cred.response as AuthenticatorAssertionResponse).userHandle!) : null,
            },
        };

        const finishRes = await fetch(`/api/auth/login/finish/${username}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
        });
        if (!finishRes.ok) throw new Error(await finishRes.text());
    }
};

// Utils
function base64URLToBuffer(base64URL: string): ArrayBuffer {
    const base64 = base64URL.replace(/-/g, '+').replace(/_/g, '/');
    const padLen = (4 - (base64.length % 4)) % 4;
    return Uint8Array.from(atob(base64 + '='.repeat(padLen)), c => c.charCodeAt(0)).buffer;
}

function bufferToBase64URL(buffer: ArrayBuffer): string {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}
