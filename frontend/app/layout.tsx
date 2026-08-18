import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "AgriAgent — Smart Farming Dashboard",
  description: "ระบบ Smart Farming สำหรับฟาร์มมะเขือเทศออร์แกนิก",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="th" className="h-full antialiased">
      <head>
        <link
          href="https://fonts.googleapis.com/css2?family=Be+Vietnam+Pro:wght@400;600;700&display=swap"
          rel="stylesheet"
        />
        <link
          href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="min-h-full">{children}</body>
    </html>
  );
}
