import Sidebar from "@/components/Sidebar";
import BottomNav from "@/components/BottomNav";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <div className="flex-1 md:ml-72">
        {children}
      </div>
      <BottomNav />
    </div>
  );
}
