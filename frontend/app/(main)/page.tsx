"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import MobileHeader from "@/components/MobileHeader";
import TrendChart from "@/components/TrendChart";
import { api } from "@/lib/api";
import type { AgentDecision, LeafImage, ReadingsRange, SensorReading } from "@/lib/types";

const POLL_MS = 15000;

const CV_COLOR: Record<string, string> = {
  healthy: "#66bb6a",
  yellowing: "#ffa726",
};

const ACTION_LABEL: Record<string, string> = {
  water_on: "เปิดน้ำ",
  water_off: "ปิดน้ำ",
  no_action: "ไม่ต้องทำอะไร",
  alert: "แจ้งเตือน",
};

const METRIC_TABS: { key: "soil_moisture" | "temperature" | "humidity"; label: string }[] = [
  { key: "soil_moisture", label: "ความชื้นดิน" },
  { key: "temperature", label: "อุณหภูมิ" },
  { key: "humidity", label: "ความชื้นอากาศ" },
];

const RANGE_TABS: { key: ReadingsRange; label: string }[] = [
  { key: "day", label: "วัน" },
  { key: "month", label: "เดือน" },
  { key: "year", label: "ปี" },
];

function timeAgo(iso: string) {
  const mins = Math.round((Date.now() - new Date(iso).getTime()) / 60000);
  if (mins < 1) return "เมื่อสักครู่";
  if (mins < 60) return `${mins} นาทีที่แล้ว`;
  const hrs = Math.round(mins / 60);
  if (hrs < 24) return `${hrs} ชม.ที่แล้ว`;
  return `${Math.round(hrs / 24)} วันที่แล้ว`;
}

export default function DashboardPage() {
  const [latest, setLatest] = useState<SensorReading[]>([]);
  const [rangeData, setRangeData] = useState<SensorReading[]>([]);
  const [range, setRange] = useState<ReadingsRange>("day");
  const [metric, setMetric] = useState<"soil_moisture" | "temperature" | "humidity">("soil_moisture");
  const [images, setImages] = useState<LeafImage[]>([]);
  const [decisions, setDecisions] = useState<AgentDecision[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);

  const loadAll = useCallback(async (r: ReadingsRange) => {
    try {
      const [latestData, rangeResult, imgData, decisionData] = await Promise.all([
        api.latestReadings(),
        api.readingsRange(r),
        api.recentImages(6),
        api.decisions(),
      ]);
      setLatest(latestData);
      setRangeData(rangeResult);
      setImages(imgData);
      setDecisions(decisionData);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "โหลดข้อมูลไม่สำเร็จ");
    }
  }, []);

  useEffect(() => {
    function tick() {
      loadAll(range);
    }
    tick();
    const id = setInterval(tick, POLL_MS);
    return () => clearInterval(id);
  }, [range, loadAll]);

  async function handleConfirm(id: number, approve: boolean) {
    setBusyId(id);
    try {
      const updated = await api.confirmDecision(id, approve);
      setDecisions((prev) => prev.map((d) => (d.id === id ? updated : d)));
    } finally {
      setBusyId(null);
    }
  }

  const primaryReading = latest[0];
  const soilPct = primaryReading ? Math.round(Math.min(100, Math.max(0, primaryReading.soil_moisture))) : null;
  const tempC = primaryReading ? Math.round(primaryReading.temperature) : null;
  const tempRingPct = primaryReading ? Math.min(100, Math.max(0, (primaryReading.temperature / 40) * 100)) : 0;

  const latestImage = images[0];
  const healthScore = latestImage
    ? latestImage.cv_result === "healthy"
      ? Math.round(latestImage.confidence * 100)
      : Math.round((1 - latestImage.confidence) * 100)
    : null;

  const pendingDecisions = decisions.filter((d) => d.status === "pending");
  const deviceOptions = useMemo(
    () => Array.from(new Set(rangeData.map((r) => r.device_id))),
    [rangeData]
  );
  const chartDevice = deviceOptions[0] ?? "";

  return (
    <div className="pb-24 md:pb-8 pt-20 md:pt-8 min-h-screen">
      <MobileHeader title="AgriAgent" />
      <main className="w-full max-w-[1200px] mx-auto px-4 md:px-8 flex flex-col gap-5">
        <section className="flex flex-col gap-2 pt-4">
          <h1 className="text-2xl md:text-3xl font-semibold text-on-surface">แดชบอร์ดฟาร์ม</h1>

          {error && (
            <div className="bg-error-container text-on-error-container rounded-lg p-3 text-[14px]">
              โหลดข้อมูลไม่สำเร็จ: {error} — ตรวจสอบว่า backend รันอยู่ที่ NEXT_PUBLIC_API_BASE_URL
            </div>
          )}

          {pendingDecisions.length > 0 ? (
            <div className="bg-error-container/40 text-on-error-container rounded-xl p-4 flex items-start gap-4 soft-shadow">
              <span className="material-symbols-outlined fill mt-1 text-error">notification_important</span>
              <div>
                <h3 className="text-[20px] leading-7 font-semibold">รอการยืนยันจากคุณ</h3>
                <p className="text-[14px] leading-5 mt-1 opacity-90">
                  มี {pendingDecisions.length} รายการที่ agent ไม่มั่นใจพอ ต้องการให้คุณยืนยันก่อนทำงาน
                </p>
              </div>
            </div>
          ) : (
            <div className="bg-primary-container text-on-primary-container rounded-xl p-4 flex items-start gap-4 soft-shadow">
              <span className="material-symbols-outlined fill mt-1" style={{ color: "#95d4b3" }}>
                workspace_premium
              </span>
              <div>
                <h3 className="text-[20px] leading-7 font-semibold">ดูแลได้ดีเยี่ยม!</h3>
                <p className="text-[14px] leading-5 mt-1 opacity-90">ไม่มีรายการที่ต้องยืนยัน ระบบทำงานอัตโนมัติตามปกติ</p>
              </div>
            </div>
          )}
        </section>

        <section className="grid grid-cols-1 md:grid-cols-3 gap-5">
          <div className="bg-surface rounded-xl p-6 soft-shadow flex flex-col items-center justify-center relative">
            <div
              className="w-32 h-32 rounded-full ring-temp flex items-center justify-center relative mb-4"
              style={{ "--val": `${tempRingPct}%` } as React.CSSProperties}
            >
              <div className="w-28 h-28 bg-surface rounded-full flex flex-col items-center justify-center">
                <span className="material-symbols-outlined fill mb-1" style={{ color: "#ffa726" }}>
                  thermostat
                </span>
                <span className="text-3xl font-bold text-on-surface">{tempC ?? "-"}°C</span>
              </div>
            </div>
            <h3 className="text-[20px] font-semibold text-on-surface mb-2">อุณหภูมิ</h3>
            <div className="bg-surface-container-low px-3 py-1 rounded-full flex items-center gap-1">
              <span className="text-outline text-[14px]">{primaryReading?.device_id ?? "ไม่มีข้อมูล"}</span>
            </div>
          </div>

          <div className="bg-surface rounded-xl p-6 soft-shadow flex flex-col items-center justify-center relative">
            <div
              className="w-32 h-32 rounded-full ring-moisture flex items-center justify-center relative mb-4"
              style={{ "--val": `${soilPct ?? 0}%` } as React.CSSProperties}
            >
              <div className="w-28 h-28 bg-surface rounded-full flex flex-col items-center justify-center">
                <span className="material-symbols-outlined fill mb-1" style={{ color: "#42a5f5" }}>
                  water_drop
                </span>
                <span className="text-3xl font-bold text-on-surface">{soilPct ?? "-"}%</span>
              </div>
            </div>
            <h3 className="text-[20px] font-semibold text-on-surface mb-2">ความชื้นดิน</h3>
            <div className="bg-surface-container-low px-3 py-1 rounded-full flex items-center gap-1">
              <span className="text-outline text-[14px]">
                {primaryReading ? timeAgo(primaryReading.created_at) : "ไม่มีข้อมูล"}
              </span>
            </div>
          </div>

          <div className="bg-surface rounded-xl p-6 soft-shadow flex flex-col items-center justify-center relative">
            <div
              className="w-32 h-32 rounded-full ring-health flex items-center justify-center relative mb-4"
              style={{ "--val": `${healthScore ?? 0}%` } as React.CSSProperties}
            >
              <div className="w-28 h-28 bg-surface rounded-full flex flex-col items-center justify-center">
                <span className="material-symbols-outlined fill mb-1" style={{ color: "#66bb6a" }}>
                  health_and_safety
                </span>
                <span className="text-3xl font-bold text-on-surface">{healthScore ?? "-"}</span>
              </div>
            </div>
            <h3 className="text-[20px] font-semibold text-on-surface mb-2">คะแนนสุขภาพใบ</h3>
            <div className="bg-surface-container-low px-3 py-1 rounded-full flex items-center gap-1">
              <span className="text-outline text-[14px]">
                {latestImage ? (latestImage.cv_result === "healthy" ? "ปกติ" : latestImage.cv_result) : "ไม่มีภาพ"}
              </span>
            </div>
          </div>
        </section>

        {pendingDecisions.slice(0, 3).map((d) => (
          <section
            key={d.id}
            className="bg-tertiary-container rounded-xl p-6 floating-shadow flex flex-col md:flex-row items-start md:items-center justify-between gap-4"
          >
            <div className="flex items-start gap-4">
              <div className="bg-on-tertiary p-3 rounded-full flex-shrink-0" style={{ background: "rgba(255,255,255,0.2)" }}>
                <span className="material-symbols-outlined fill text-on-tertiary">smart_toy</span>
              </div>
              <div>
                <h3 className="text-[20px] font-semibold text-on-tertiary mb-1">
                  คำแนะนำจาก AI · {d.device_id} · {ACTION_LABEL[d.action] ?? d.action}
                </h3>
                <p className="text-[14px] text-on-tertiary-container">{d.reason}</p>
                <p className="text-[12px] text-on-tertiary-container/70 mt-1">
                  ความมั่นใจ {(d.confidence * 100).toFixed(0)}%
                </p>
              </div>
            </div>
            <div className="flex gap-2 w-full md:w-auto mt-2 md:mt-0">
              <button
                disabled={busyId === d.id}
                onClick={() => handleConfirm(d.id, true)}
                className="flex-1 md:flex-none bg-surface-container-lowest text-tertiary text-[12px] font-bold tracking-wider uppercase px-6 py-3 rounded-full hover:bg-surface-container-high transition-colors disabled:opacity-50"
              >
                ยืนยัน
              </button>
              <button
                disabled={busyId === d.id}
                onClick={() => handleConfirm(d.id, false)}
                className="flex-1 md:flex-none border border-outline-variant text-on-tertiary text-[12px] font-bold tracking-wider uppercase px-6 py-3 rounded-full hover:bg-white/10 transition-colors disabled:opacity-50"
              >
                ไม่ใช่
              </button>
            </div>
          </section>
        ))}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
          <section className="lg:col-span-2 bg-surface rounded-xl p-6 soft-shadow flex flex-col">
            <div className="flex flex-wrap justify-between items-center gap-3 mb-4">
              <div className="flex gap-1 bg-surface-container-low rounded-full p-1">
                {METRIC_TABS.map((m) => (
                  <button
                    key={m.key}
                    onClick={() => setMetric(m.key)}
                    className={`px-3 py-1.5 rounded-full text-[12px] font-bold tracking-wider uppercase transition-colors ${
                      metric === m.key ? "bg-primary text-on-primary shadow-sm" : "text-on-surface-variant hover:bg-surface-variant"
                    }`}
                  >
                    {m.label}
                  </button>
                ))}
              </div>
              <div className="bg-surface-container-low rounded-full p-1 flex">
                {RANGE_TABS.map((r) => (
                  <button
                    key={r.key}
                    onClick={() => setRange(r.key)}
                    className={`px-4 py-1.5 rounded-full text-[12px] font-bold tracking-wider uppercase transition-colors ${
                      range === r.key ? "bg-primary text-on-primary shadow-sm" : "text-on-surface-variant hover:bg-surface-variant"
                    }`}
                  >
                    {r.label}
                  </button>
                ))}
              </div>
            </div>
            {chartDevice ? (
              <TrendChart data={rangeData} deviceId={chartDevice} metric={metric} range={range} />
            ) : (
              <div className="flex h-[220px] items-center justify-center text-[14px] text-outline">กำลังโหลดข้อมูล...</div>
            )}
          </section>

          <section className="bg-surface rounded-xl p-6 soft-shadow">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-[20px] font-semibold text-on-surface">Plant Log</h3>
              <span className="text-outline text-[12px] font-bold tracking-wider uppercase">ทุก 3 ชม.</span>
            </div>
            <div className="flex flex-col gap-3">
              {images.length === 0 && <p className="text-[14px] text-on-surface-variant">ยังไม่มีภาพ log</p>}
              {images.map((img) => (
                <div key={img.id} className="flex items-center gap-3 p-2 rounded-lg hover:bg-surface-container-low transition-colors">
                  <div className="w-12 h-12 rounded-lg overflow-hidden flex-shrink-0 relative">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={img.image_url} alt={img.cv_result} className="w-full h-full object-cover" />
                    <div
                      className="absolute top-1 right-1 w-2.5 h-2.5 rounded-full border border-white"
                      style={{ background: CV_COLOR[img.cv_result] ?? "#ef5350" }}
                    />
                  </div>
                  <div className="flex-1">
                    <p className="text-[14px] font-semibold text-on-surface">{img.device_id}</p>
                    <p className="text-[12px] font-bold tracking-wider uppercase text-outline">
                      {img.cv_result === "healthy" ? "ปกติ" : img.cv_result} · {timeAgo(img.created_at)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}
