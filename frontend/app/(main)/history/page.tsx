"use client";

import { useEffect, useMemo, useState } from "react";
import MobileHeader from "@/components/MobileHeader";
import { api } from "@/lib/api";
import type { AgentDecision, ReadingsRange } from "@/lib/types";

const POLL_MS = 15000;

const ACTION_LABEL: Record<string, string> = {
  water_on: "เปิดน้ำ",
  water_off: "ปิดน้ำ",
  no_action: "ไม่ต้องทำอะไร",
  alert: "แจ้งเตือน",
};

const RANGE_TABS: { key: ReadingsRange; label: string }[] = [
  { key: "day", label: "วัน" },
  { key: "month", label: "เดือน" },
  { key: "year", label: "ปี" },
];

type Category = "agent" | "alert" | "user";

function categorize(d: AgentDecision): Category {
  if (d.action === "alert") return "alert";
  if (d.confirmed_by) return "user";
  return "agent";
}

const CATEGORY_META: Record<Category, { icon: string; iconBg: string; iconText: string; tag: string; tagBg: string }> = {
  agent: { icon: "smart_toy", iconBg: "bg-tertiary-container", iconText: "text-on-tertiary-container", tag: "Automation", tagBg: "bg-tertiary-container/20 text-tertiary" },
  user: { icon: "person", iconBg: "bg-primary-container", iconText: "text-on-primary-container", tag: "Manual", tagBg: "bg-primary-container/20 text-primary" },
  alert: { icon: "warning", iconBg: "bg-error-container", iconText: "text-on-error-container", tag: "Alert", tagBg: "bg-error-container/50 text-error" },
};

function within(iso: string, range: ReadingsRange) {
  const now = Date.now();
  const t = new Date(iso).getTime();
  const day = 24 * 60 * 60 * 1000;
  if (range === "day") return now - t <= day;
  if (range === "month") return now - t <= 30 * day;
  return now - t <= 365 * day;
}

export default function HistoryPage() {
  const [decisions, setDecisions] = useState<AgentDecision[]>([]);
  const [range, setRange] = useState<ReadingsRange>("month");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const res = await api.decisions();
        if (!cancelled) {
          setDecisions(res);
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

  const filtered = useMemo(
    () => decisions.filter((d) => within(d.created_at, range)),
    [decisions, range]
  );

  const grouped = useMemo(() => {
    const groups = new Map<string, AgentDecision[]>();
    for (const d of filtered) {
      const key = new Date(d.created_at).toLocaleDateString("th-TH", { day: "numeric", month: "long", year: "numeric" });
      const list = groups.get(key) ?? [];
      list.push(d);
      groups.set(key, list);
    }
    return Array.from(groups.entries());
  }, [filtered]);

  return (
    <div className="pb-24 md:pb-8 pt-20 md:pt-8 min-h-screen">
      <MobileHeader title="รายงาน" />
      <main className="w-full max-w-[1024px] mx-auto px-4 md:px-8 flex flex-col gap-12 py-8">
        <section className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
          <div>
            <h1 className="text-2xl md:text-3xl font-semibold text-on-surface mb-2">ประวัติการทำงาน</h1>
            <p className="text-[16px] leading-6 text-on-surface-variant">บันทึกกิจกรรมและการแจ้งเตือนระบบทั้งหมด</p>
          </div>
          <div className="inline-flex bg-surface-container rounded-full p-1 soft-shadow">
            {RANGE_TABS.map((r) => (
              <button
                key={r.key}
                onClick={() => setRange(r.key)}
                className={`px-6 py-2 rounded-full text-[12px] font-bold tracking-wider uppercase transition-all ${
                  range === r.key ? "bg-primary text-on-primary shadow-sm" : "text-on-surface-variant hover:bg-surface-variant"
                }`}
              >
                {r.label}
              </button>
            ))}
          </div>
        </section>

        {error && (
          <div className="bg-error-container text-on-error-container rounded-lg p-3 text-[14px]">
            โหลดข้อมูลไม่สำเร็จ: {error}
          </div>
        )}

        {grouped.length === 0 && !error && (
          <p className="text-[14px] text-on-surface-variant">ไม่มีประวัติในช่วงเวลานี้</p>
        )}

        {grouped.map(([day, items]) => (
          <section key={day} className="relative w-full max-w-4xl mx-auto timeline-container">
            <div className="relative z-10 flex items-center gap-4 mb-8">
              <div className="bg-surface-container-high text-on-surface-variant px-4 py-1 rounded-full text-[12px] font-bold tracking-wider uppercase border border-outline-variant inline-block">
                {day}
              </div>
            </div>
            {items.map((d) => {
              const cat = categorize(d);
              const meta = CATEGORY_META[cat];
              const title = d.confirmed_by ?? "AI Agent";
              const desc = `${ACTION_LABEL[d.action] ?? d.action} · ${d.device_id} — ${d.reason}`;
              const time = new Date(d.created_at).toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" });
              return (
                <div key={d.id} className="relative z-10 flex gap-6 mb-8 group">
                  <div
                    className={`w-12 h-12 flex-shrink-0 rounded-full ${meta.iconBg} flex items-center justify-center soft-shadow border-4 border-background group-hover:scale-110 transition-transform duration-200`}
                  >
                    <span className={`material-symbols-outlined ${meta.iconText} fill`}>{meta.icon}</span>
                  </div>
                  <div
                    className={`flex-1 bg-surface-container-lowest rounded-xl p-6 soft-shadow hover:shadow-[0px_8px_30px_rgba(0,0,0,0.12)] transition-shadow ${cat === "alert" ? "border border-error-container/30" : ""}`}
                  >
                    <div className="flex justify-between items-start mb-2">
                      <h3 className="text-[20px] font-semibold text-on-surface flex items-center gap-2">
                        {title}
                        <span className={`px-2 py-0.5 rounded-full ${meta.tagBg} text-[12px] font-bold tracking-wider uppercase`}>
                          {meta.tag}
                        </span>
                        {d.status === "pending" && (
                          <span className="px-2 py-0.5 rounded-full bg-error-container text-on-error-container text-[12px] font-bold tracking-wider uppercase">
                            รอยืนยัน
                          </span>
                        )}
                      </h3>
                      <span className={`text-[14px] leading-5 ${cat === "alert" ? "text-error" : "text-outline"}`}>{time}</span>
                    </div>
                    <p className="text-[16px] leading-6 text-on-surface-variant">{desc}</p>
                  </div>
                </div>
              );
            })}
          </section>
        ))}
      </main>
    </div>
  );
}
