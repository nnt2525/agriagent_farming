import Link from "next/link";

export default function RegisterPage() {
  return (
    <div className="bg-surface-bright min-h-screen flex flex-col justify-center items-center p-4 relative overflow-hidden">
      <div className="fixed top-0 left-0 w-full h-full overflow-hidden -z-10 pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] rounded-full bg-primary-fixed opacity-30 blur-3xl mix-blend-multiply" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] rounded-full bg-tertiary-fixed opacity-20 blur-3xl mix-blend-multiply" />
      </div>
      <div className="w-full max-w-md bg-surface-container-lowest rounded-xl p-8 relative z-10" style={{boxShadow:'0px 4px 20px rgba(0,0,0,0.05)'}}>
        <div className="text-center mb-8">
          <h1 className="text-5xl font-bold tracking-tight text-primary mb-2">AgriAgent</h1>
          <p className="text-[16px] leading-6 text-on-surface-variant">สร้างบัญชีเพื่อเริ่มต้นใช้งาน</p>
        </div>
        <form className="space-y-6">
          {[
            { id: 'fullname', label: 'ชื่อ-นามสกุล', icon: 'person', type: 'text', placeholder: 'กรอกชื่อ-นามสกุล' },
            { id: 'email', label: 'อีเมล', icon: 'mail', type: 'email', placeholder: 'กรอกอีเมล' },
            { id: 'farmname', label: 'ชื่อฟาร์ม', icon: 'agriculture', type: 'text', placeholder: 'กรอกชื่อฟาร์ม' },
            { id: 'password', label: 'รหัสผ่าน', icon: 'lock', type: 'password', placeholder: 'กรอกรหัสผ่าน' },
          ].map(({ id, label, icon, type, placeholder }) => (
            <div key={id}>
              <label className="block text-[20px] font-semibold text-on-surface mb-2" htmlFor={id}>{label}</label>
              <div className="relative">
                <span className="material-symbols-outlined absolute left-3 top-1/2 transform -translate-y-1/2 text-outline">{icon}</span>
                <input id={id} name={id} type={type} placeholder={placeholder}
                  className="w-full pl-10 pr-4 py-3 bg-surface border border-outline-variant rounded-lg focus:ring-2 focus:ring-primary focus:border-primary text-[16px] text-on-surface placeholder:text-outline-variant" />
              </div>
            </div>
          ))}
          <button type="submit" className="w-full py-3 bg-primary text-on-primary rounded-full text-[20px] font-semibold hover:bg-primary-container transition-colors shadow-sm">
            สร้างบัญชี
          </button>
        </form>
        <div className="mt-8 text-center">
          <p className="text-[16px] leading-6 text-on-surface-variant">
            มีบัญชีอยู่แล้ว? <Link href="/login" className="text-primary hover:text-primary-container font-semibold transition-colors underline">เข้าสู่ระบบ</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
