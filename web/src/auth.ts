export interface User {
  id: string
  email: string
  role: string
}

interface AuthResponse extends User {
  token: string
}

let currentUser: User | null | undefined

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })

  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? 'Request failed')
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}

export async function login(email: string, password: string) {
  const response = await request<AuthResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  currentUser = response
  return response
}

export async function register(email: string, password: string) {
  const response = await request<AuthResponse>('/api/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  currentUser = response
  return response
}

export async function me() {
  try {
    currentUser = await request<User>('/api/auth/me')
  } catch {
    currentUser = null
  }
  return currentUser
}

export async function isAuthenticated() {
  return (await me()) !== null
}

export async function logout() {
  await request<void>('/api/auth/logout', { method: 'POST' })
  currentUser = null
}