"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  { href: "/", icon: "dashboard", label: "แดชบอร์ด" },
  { href: "/devices", icon: "sensors", label: "อุปกรณ์" },
  { href: "/history", icon: "analytics", label: "รายงาน" },
];

export default function Sidebar() {
  const pathname = usePathname();
  return (
    <aside className="hidden md:flex fixed top-0 left-0 w-72 h-full bg-surface z-50 flex-col p-6 border-r border-outline-variant/20" style={{boxShadow:'4px 0 24px rgba(0,0,0,0.03)'}}>
      <div className="text-4xl font-bold text-primary mb-10 pl-2">AgriAgent</div>
      <nav className="flex flex-col gap-2 flex-1">
        {navItems.map(({ href, icon, label }) => {
          const active = pathname === href;
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-3 rounded-xl px-4 py-3 transition-all ${
                active
                  ? "bg-primary-container text-on-primary-container soft-shadow"
                  : "text-on-surface-variant hover:bg-surface-container hover:text-on-surface"
              }`}
            >
              <span className={`material-symbols-outlined ${active ? 'fill' : ''}`}>{icon}</span>
              <span className="text-[20px] leading-7 font-semibold">{label}</span>
            </Link>
          );
        })}
      </nav>
      <div className="mt-auto border-t border-outline-variant/30 pt-6 flex flex-col gap-4">
        <div className="flex items-center justify-between px-2">
          <button className="text-on-surface-variant hover:text-primary hover:bg-surface-container-high transition-colors p-2 rounded-full">
            <span className="material-symbols-outlined">notifications</span>
          </button>
          <button className="text-on-surface-variant hover:text-primary hover:bg-surface-container-high transition-colors p-2 rounded-full">
            <span className="material-symbols-outlined">settings</span>
          </button>
        </div>
        <div className="flex items-center gap-3 px-2">
          <div className="w-10 h-10 rounded-full bg-primary-container flex items-center justify-center flex-shrink-0 border-2 border-primary">
            <span className="material-symbols-outlined text-on-primary-container fill">person</span>
          </div>
          <div className="overflow-hidden">
            <p className="text-[14px] font-semibold truncate text-on-surface">คุณสมชาย</p>
            <p className="text-[10px] tracking-wider font-bold uppercase text-outline truncate">ผู้ดูแลระบบ</p>
          </div>
        </div>
      </div>
    </aside>
  );
}
