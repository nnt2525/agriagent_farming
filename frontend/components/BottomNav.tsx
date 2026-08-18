"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  { href: "/", icon: "dashboard", label: "แดชบอร์ด" },
  { href: "/devices", icon: "sensors", label: "อุปกรณ์" },
  { href: "/history", icon: "analytics", label: "รายงาน" },
];

export default function BottomNav() {
  const pathname = usePathname();
  return (
    <nav className="md:hidden fixed bottom-0 left-0 w-full z-50 flex justify-around items-center px-4 pb-4 pt-2 bg-surface shadow-[0px_-4px_20px_rgba(0,0,0,0.05)] rounded-t-xl border-t border-outline-variant/30">
      {navItems.map(({ href, icon, label }) => {
        const active = pathname === href;
        return (
          <Link
            key={href}
            href={href}
            className={`flex flex-col items-center justify-center transition-all rounded-full px-5 py-1.5 ${
              active
                ? "bg-primary-container text-on-primary-container"
                : "text-outline hover:bg-surface-container-highest"
            }`}
          >
            <span className={`material-symbols-outlined mb-1 ${active ? 'fill' : ''}`}>{icon}</span>
            <span className="text-[12px] font-bold tracking-wider uppercase">{label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
