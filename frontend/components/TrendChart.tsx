"use client";

import { useMemo } from "react";
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { ReadingsRange, SensorReading } from "@/lib/types";

const METRIC_LABEL: Record<string, string> = {
  soil_moisture: "ความชื้นในดิน",
  temperature: "อุณหภูมิ",
  humidity: "ความชื้นในอากาศ",
};

const METRIC_UNIT: Record<string, string> = {
  soil_moisture: "%",
  temperature: "°C",
  humidity: "%",
};

function formatTick(iso: string, range: ReadingsRange) {
  const d = new Date(iso);
  if (range === "day") return d.toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" });
  if (range === "month") return d.toLocaleDateString("th-TH", { day: "2-digit", month: "2-digit" });
  return d.toLocaleDateString("th-TH", { month: "short" });
}

export default function TrendChart({
  data,
  deviceId,
  metric,
  range,
}: {
  data: SensorReading[];
  deviceId: string;
  metric: "soil_moisture" | "temperature" | "humidity";
  range: ReadingsRange;
}) {
  const filtered = useMemo(
    () =>
      data
        .filter((r) => r.device_id === deviceId)
        .sort((a, b) => a.created_at.localeCompare(b.created_at)),
    [data, deviceId]
  );

  if (filtered.length === 0) {
    return (
      <div className="flex h-[220px] items-center justify-center text-[14px] text-outline">
        ยังไม่มีข้อมูลย้อนหลังสำหรับช่วงนี้
      </div>
    );
  }

  return (
    <div style={{ height: 220 }}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={filtered} margin={{ top: 8, right: 8, bottom: 0, left: -20 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--color-outline-variant)" vertical={false} />
          <XAxis
            dataKey="created_at"
            tickFormatter={(v) => formatTick(v, range)}
            tick={{ fill: "var(--color-outline)", fontSize: 11 }}
            stroke="var(--color-outline-variant)"
            minTickGap={32}
          />
          <YAxis
            tick={{ fill: "var(--color-outline)", fontSize: 11 }}
            stroke="var(--color-outline-variant)"
            width={34}
            domain={metric === "temperature" ? ["auto", "auto"] : [0, 100]}
          />
          <Tooltip
            contentStyle={{
              background: "var(--color-surface-container-lowest)",
              border: "1px solid var(--color-outline-variant)",
              borderRadius: 8,
              fontSize: 12,
              color: "var(--color-on-surface)",
            }}
            labelFormatter={(v) => new Date(v as string).toLocaleString("th-TH")}
            formatter={(value) => [`${Number(value).toFixed(1)}${METRIC_UNIT[metric]}`, METRIC_LABEL[metric]]}
          />
          <Line
            type="monotone"
            dataKey={metric}
            stroke="var(--color-primary)"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4 }}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
