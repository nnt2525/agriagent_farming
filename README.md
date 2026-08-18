# AgriAgent — Smart Farming (hackathon prototype)

ระบบ Smart Farming สำหรับฟาร์มมะเขือเทศออร์แกนิก: รับข้อมูลเซนเซอร์จาก ESP32 ผ่าน MQTT, ให้ agent (LLM) ตัดสินใจรดน้ำแบบ event-driven พร้อม human-in-the-loop, และแดชบอร์ด Next.js สำหรับติดตาม/ยืนยันผล

## โครงสร้างโปรเจกต์

```
/backend   Go + Fiber API, MQTT subscriber, agent service (migrations/ มี SQL schema)
/frontend  Next.js (App Router) dashboard
/model     ไฟล์ CV model (leaf disease detection) — ยังไม่ได้ต่อเข้ากับ backend
/img, /img_output  ตัวอย่างภาพสำหรับทดสอบ CV model
```

## เริ่มรัน backend

```bash
cd backend
cp .env.example .env      # ปรับค่าตามจริง; ค่า default รันได้เลยแบบ mock mode
go run ./cmd/server
```

- ค่า default: `MOCK_MODE=true` → API คืนข้อมูล mock ที่ seed มาให้ (ไม่ต้องมี Postgres/MQTT ก็ทดสอบ frontend ได้ทันที)
- ต่อ Postgres/Supabase จริง: รัน `backend/migrations/0001_init.sql` กับ DB ก่อน แล้วตั้ง `MOCK_MODE=false` และ `DATABASE_URL`
- ต่อ MQTT จริง (HiveMQ Cloud หรือ broker อื่นที่รองรับ TLS): ตั้งค่า `MQTT_BROKER_URL`, `MQTT_USERNAME`, `MQTT_PASSWORD`
- ต่อ LLM จริง (Gemini): ตั้งค่า `GEMINI_API_KEY` — ถ้าไม่ตั้งจะ fallback เป็น rule-based mock decision (ใช้ demo ได้โดยไม่ต้องมี API key)
- Backend ฟังที่ `:8080` โดย default

## เริ่มรัน frontend

```bash
cd frontend
cp .env.example .env.local   # NEXT_PUBLIC_API_BASE_URL ชี้ไปที่ backend
npm install
npm run dev
```

เปิด http://localhost:3000 — มี 3 หน้าหลัก (แดชบอร์ด, อุปกรณ์, ประวัติ) บวกหน้า login/register แบบ static (ยังไม่ต่อ auth เพราะ backend ไม่มีระบบ auth ตามสโคปปัจจุบัน)

## Agent behavior (event-driven)

Trigger การประเมินผลของ agent เกิดขึ้นเมื่อ:
- soil_moisture ต่ำกว่า `SOIL_MOISTURE_LOW_PCT` (default 30%)
- soil_moisture เปลี่ยนกะทันหันเกิน `SOIL_MOISTURE_DELTA_PCT` (default 10%)
- ครบรอบ `AGENT_SCHEDULE_INTERVAL` (default 3 ชม.)
- เรียกผ่าน `POST /api/agent/trigger` (ใช้ตอน demo)
- CV เจอใบผิดปกติ (ผ่าน trigger source `cv` — ต้องมีระบบเขียนผล CV ลง `leaf_images` ก่อน ดู `model/` ที่ยังไม่ได้เชื่อมกับ backend)

ผลตัดสินใจ (`action`, `reason`, `confidence`, `need_human_confirm`) ถูกบันทึกลง `agent_decisions` เสมอ — ถ้ามั่นใจสูง (>=0.85 ตาม prompt) จะ auto-execute, ถ้าไม่มั่นใจจะรอ `POST /api/decisions/:id/confirm` จากแอดมิน

## สถานะปัจจุบัน / สิ่งที่ยังไม่เสร็จ

- `model/` มี CV model files (`best.pt`, `tomato_leaf_disease_model.keras`) แต่ยังไม่มี service ที่รันโมเดลแล้วเขียนผลลง `leaf_images`/trigger agent — ต้องต่อเพิ่ม
- หน้า login/register เป็น static mockup เฉยๆ ยังไม่มี backend auth
- Relay actuation (`agent.ExecuteFunc`) ยังเป็น stub แค่ log — ต้องต่อ MQTT publish ไปยัง relay topic จริงเมื่อ ESP32 relay firmware กำหนด topic แล้ว
