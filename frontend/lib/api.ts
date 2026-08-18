import type {
  AgentDecision,
  DevicesResponse,
  LeafImage,
  ReadingsRange,
  SensorReading,
} from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
    cache: "no-store",
  });
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`${init?.method ?? "GET"} ${path} failed (${res.status}): ${text}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  latestReadings: () => apiFetch<SensorReading[]>("/api/readings/latest"),
  readingsRange: (range: ReadingsRange) =>
    apiFetch<SensorReading[]>(`/api/readings?range=${range}`),
  devices: () => apiFetch<DevicesResponse>("/api/devices"),
  recentImages: (limit = 20) => apiFetch<LeafImage[]>(`/api/images/recent?limit=${limit}`),
  decisions: () => apiFetch<AgentDecision[]>("/api/decisions"),
  confirmDecision: (id: number, approve: boolean, confirmedBy = "admin") =>
    apiFetch<AgentDecision>(`/api/decisions/${id}/confirm`, {
      method: "POST",
      body: JSON.stringify({ approve, confirmed_by: confirmedBy }),
    }),
  triggerAgent: (deviceId: string) =>
    apiFetch<{ status: string; message: string }>("/api/agent/trigger", {
      method: "POST",
      body: JSON.stringify({ device_id: deviceId }),
    }),
};
