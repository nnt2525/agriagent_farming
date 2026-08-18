"use client";

import { useEffect, useState } from "react";
import MobileHeader from "@/components/MobileHeader";
import { api } from "@/lib/api";
import type { DevicesResponse } from "@/lib/types";

const POLL_MS = 15000;

const TYPE_ICON: Record<string, string> = {
  soil_sensor: "humidity_percentage",
  camera: "videocam",
  relay: "valve",
};

function timeAgo(iso: string | null) {
  if (!iso) return "ไม่เคยเห็น";
  const mins = Math.round((Date.now() - new Date(iso).getTime()) / 60000);
  if (mins < 1) return "เมื่อสักครู่";
  if (mins < 60) return `${mins} นาทีที่แล้ว`;
  const hrs = Math.round(mins / 60);
  if (hrs < 24) return `${hrs} ชม. ที่แล้ว`;
  return `${Math.round(hrs / 24)} วันที่แล้ว`;
}

export default function DevicesPage() {
  const [data, setData] = useState<DevicesResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const res = await api.devices();
        if (!cancelled) {
          setData(res);
          setError(null);
        }
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "โหลดข้อมูลไม่สำเร็จ");
      }
    }
    load();
    const id = setInterval(load, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  const devices = data?.devices ?? [];
  const online = data?.active ?? 0;
  const offline = data?.inactive ?? 0;

  return (
    <div className="pb-24 md:pb-8 pt-20 md:pt-8 min-h-screen">
      <MobileHeader title="อุปกรณ์" />
      <main className="w-full max-w-[1200px] mx-auto px-4 md:px-8 flex flex-col gap-5">
        {error && (
          <div className="bg-error-container text-on-error-container rounded-lg p-3 text-[14px] mt-4">
            โหลดข้อมูลไม่สำเร็จ: {error}
          </div>
        )}

        <section className="grid grid-cols-3 gap-5 pt-4">
          {[
            { label: "ทั้งหมด", value: devices.length, icon: "sensors", color: "text-on-surface" },
            { label: "เชื่อมต่อ", value: online, icon: "wifi", color: "text-primary" },
            { label: "ขาดการเชื่อมต่อ", value: offline, icon: "wifi_off", color: "text-error" },
          ].map(({ label, value, icon, color }) => (
            <div key={label} className="bg-surface-container-lowest rounded-xl p-6 soft-shadow flex flex-col gap-2">
              <div className="flex justify-between items-start">
                <span className="text-[14px] leading-5 text-on-surface-variant">{label}</span>
                <span className={`material-symbols-outlined ${color}`}>{icon}</span>
              </div>
              <div className={`text-3xl font-semibold ${color}`}>{value}</div>
            </div>
          ))}
        </section>

        {offline > 0 && (
          <section>
            <div className="bg-error-container text-on-error-container rounded-lg p-4 flex items-center gap-3">
              <span className="material-symbols-outlined text-error fill">error</span>
              <span className="text-[16px] leading-6">ตรวจพบอุปกรณ์ขาดการเชื่อมต่อ {offline} รายการ</span>
            </div>
          </section>
        )}

        <section className="flex flex-col gap-4">
          <h2 className="text-[20px] font-semibold text-on-surface">รายการอุปกรณ์</h2>
          <div className="flex flex-col gap-3">
            {devices.map((d) => {
              const isOnline = d.status === "active";
              return (
                <div
                  key={d.id}
                  className={`bg-surface-container-lowest rounded-xl p-4 soft-shadow flex items-center justify-between ${!isOnline ? "opacity-75" : ""}`}
                >
                  <div className="flex items-center gap-4">
                    <div
                      className={`w-12 h-12 rounded-full bg-surface-container-low flex items-center justify-center ${isOnline ? "text-primary" : "text-outline"}`}
                    >
                      <span className="material-symbols-outlined">{TYPE_ICON[d.type] ?? "device_unknown"}</span>
                    </div>
                    <div className="flex flex-col">
                      <span className="text-[20px] font-semibold text-on-surface">{d.name}</span>
                      <span className="text-[14px] leading-5 text-on-surface-variant">
                        {d.zone} • เห็นล่าสุด {timeAgo(d.last_seen)}
                      </span>
                    </div>
                  </div>
                  {isOnline ? (
                    <div className="bg-primary-container text-on-primary-container px-3 py-1 rounded-full text-[12px] font-bold tracking-wider uppercase flex items-center gap-1">
                      <span className="w-2 h-2 rounded-full bg-primary inline-block" /> ออนไลน์
                    </div>
                  ) : (
                    <div className="bg-surface-variant text-on-surface-variant px-3 py-1 rounded-full text-[12px] font-bold tracking-wider uppercase flex items-center gap-1">
                      <span className="w-2 h-2 rounded-full bg-outline inline-block" /> ออฟไลน์
                    </div>
                  )}
                </div>
              );
            })}
            {devices.length === 0 && !error && (
              <p className="text-[14px] text-on-surface-variant">ยังไม่มีอุปกรณ์</p>
            )}
          </div>
        </section>
      </main>
    </div>
  );
}
