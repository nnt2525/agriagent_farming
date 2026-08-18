export type Device = {
  id: string;
  name: string;
  type: string;
  zone: string;
  status: "active" | "inactive";
  last_seen: string | null;
};

export type DevicesResponse = {
  active: number;
  inactive: number;
  devices: Device[];
};

export type SensorReading = {
  id: number;
  device_id: string;
  soil_moisture: number;
  temperature: number;
  humidity: number;
  created_at: string;
};

export type LeafImage = {
  id: number;
  device_id: string;
  image_url: string;
  cv_result: string;
  confidence: number;
  created_at: string;
};

export type DecisionAction = "water_on" | "water_off" | "no_action" | "alert";
export type DecisionStatus = "pending" | "confirmed" | "rejected" | "auto_executed";
export type TriggerSource = "threshold" | "delta" | "cv" | "schedule" | "manual";

export type AgentDecision = {
  id: number;
  device_id: string;
  action: DecisionAction;
  reason: string;
  confidence: number;
  need_human_confirm: boolean;
  confirmed_by: string | null;
  confirmed_at: string | null;
  status: DecisionStatus;
  trigger_source: TriggerSource;
  created_at: string;
};

export type ReadingsRange = "day" | "month" | "year";
