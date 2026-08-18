export default function MobileHeader({ title }: { title: string }) {
  return (
    <header className="md:hidden fixed top-0 left-0 w-full z-50 flex justify-between items-center px-4 py-4 bg-background/80 backdrop-blur-md border-b border-surface-container">
      <span className="text-2xl leading-8 font-semibold text-primary">{title}</span>
      <div className="flex items-center gap-3">
        <button className="text-primary hover:bg-surface-container-high transition-colors p-2 rounded-full">
          <span className="material-symbols-outlined">notifications</span>
        </button>
        <div className="w-8 h-8 rounded-full bg-primary-container flex items-center justify-center border-2 border-primary-container">
          <span className="material-symbols-outlined text-on-primary-container fill" style={{fontSize:'20px'}}>person</span>
        </div>
      </div>
    </header>
  );
}
