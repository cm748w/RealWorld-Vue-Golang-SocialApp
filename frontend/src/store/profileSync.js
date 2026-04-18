const SYNC_KEY = 'rw-socialapp-profile-sync'
const CHANNEL_NAME = 'rw-socialapp-profile-sync-channel'

let channel = null
let installed = false

function isBrowser() {
    return typeof window !== 'undefined' && typeof localStorage !== 'undefined'
}

function ensureChannel() {
    if (!isBrowser() || typeof BroadcastChannel === 'undefined') {
        return null
    }

    if (!channel) {
        channel = new BroadcastChannel(CHANNEL_NAME)
    }

    return channel
}

function applyProfileSync(store, message) {
    if (!message?.type) {
        return
    }

    if (message.type === 'logout') {
        store.commit('auth/Logout', null, { root: true })
        return
    }

    const users = Array.isArray(message.payload?.users)
        ? message.payload.users
        : message.payload?.user
            ? [message.payload.user]
            : []

    users.forEach((user) => {
        if (user?._id) {
            store.commit('users/SetUser', user, { root: true })
        }
    })

    if (message.payload?.authData) {
        store.commit('auth/Auth', message.payload.authData, { root: true })
    }
}

export function installProfileSync(store) {
    if (!isBrowser() || installed) {
        return
    }

    installed = true

    const handleMessage = (message) => {
        if (!message?.data) {
            return
        }

        applyProfileSync(store, message.data)
    }

    const bc = ensureChannel()
    if (bc) {
        bc.addEventListener('message', handleMessage)
    }

    window.addEventListener('storage', (event) => {
        if (event.key !== SYNC_KEY || !event.newValue) {
            return
        }

        try {
            applyProfileSync(store, JSON.parse(event.newValue))
        } catch (error) {
            console.error('Failed to apply profile sync from storage:', error)
        }
    })
}

export function emitProfileSync(message) {
    if (!isBrowser()) {
        return
    }

    const envelope = {
        ...message,
        timestamp: Date.now(),
    }

    const bc = ensureChannel()
    if (bc) {
        bc.postMessage(envelope)
    }

    try {
        localStorage.setItem(SYNC_KEY, JSON.stringify(envelope))
    } catch (error) {
        console.error('Failed to emit profile sync:', error)
    }
}