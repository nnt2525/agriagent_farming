import Link from "next/link";

export default function LoginPage() {
  return (
    <div className="bg-background min-h-screen flex items-center justify-center p-4 md:p-8">
      <div className="w-full max-w-[440px] bg-surface-container-lowest rounded-xl p-8 md:p-10" style={{boxShadow:'0px 4px 20px rgba(0,0,0,0.05)'}}>
        <div className="flex flex-col items-center mb-8">
          <span className="material-symbols-outlined fill text-primary mb-4" style={{fontSize:'48px'}}>eco</span>
          <h1 className="text-3xl font-semibold text-primary tracking-tight">AgriAgent</h1>
          <p className="text-[14px] leading-5 text-on-surface-variant mt-2">ลงชื่อเข้าใช้เพื่อจัดการฟาร์มของคุณ</p>
        </div>
        <form className="space-y-6">
          <div className="space-y-2">
            <label className="text-[12px] font-bold tracking-wider uppercase text-on-surface-variant block" htmlFor="email">อีเมล</label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <span className="material-symbols-outlined text-outline">mail</span>
              </div>
              <input id="email" name="email" type="email" placeholder="example@farm.com" required
                className="w-full pl-10 pr-3 py-3 bg-surface border border-outline-variant rounded-lg text-[16px] text-on-surface focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary transition-colors placeholder:text-outline" />
            </div>
          </div>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-[12px] font-bold tracking-wider uppercase text-on-surface-variant block" htmlFor="password">รหัสผ่าน</label>
              <a href="#" className="text-[14px] text-primary hover:text-primary-container transition-colors">ลืมรหัสผ่าน?</a>
            </div>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <span className="material-symbols-outlined text-outline">lock</span>
              </div>
              <input id="password" name="password" type="password" placeholder="••••••••" required
                className="w-full pl-10 pr-10 py-3 bg-surface border border-outline-variant rounded-lg text-[16px] text-on-surface focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary transition-colors placeholder:text-outline" />
              <button type="button" className="absolute inset-y-0 right-0 pr-3 flex items-center text-outline hover:text-on-surface transition-colors">
                <span className="material-symbols-outlined">visibility_off</span>
              </button>
            </div>
          </div>
          <div className="flex items-center">
            <input id="remember-me" name="remember-me" type="checkbox" className="h-4 w-4 text-primary bg-surface border-outline-variant rounded focus:ring-primary focus:ring-2" />
            <label htmlFor="remember-me" className="ml-2 text-[14px] text-on-surface-variant">จดจำฉันไว้ในระบบ</label>
          </div>
          <button type="submit" className="w-full flex justify-center py-3 px-4 border border-transparent rounded-full text-[20px] font-semibold text-on-primary bg-primary hover:bg-primary-container focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary transition-colors active:scale-[0.98]">
            เข้าสู่ระบบ
          </button>
        </form>
        <div className="mt-8 text-center">
          <p className="text-[14px] text-on-surface-variant">
            ยังไม่มีบัญชีผู้ใช้?
            <Link href="/register" className="text-[20px] font-semibold text-primary hover:text-primary-container transition-colors ml-1 underline underline-offset-2">ลงทะเบียนที่นี่</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
