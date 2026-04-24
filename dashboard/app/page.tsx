import Link from "next/link"

export default function HomePage() {
  return (
    <main className="flex min-h-svh items-center justify-center bg-[linear-gradient(145deg,#fff7ed_0%,#ffffff_38%,#ecfeff_100%)] px-6 py-16">
      <div className="w-full max-w-3xl rounded-[32px] border border-white/70 bg-white/90 p-10 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur">
        <div className="inline-flex rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-xs font-medium tracking-[0.18em] text-amber-900 uppercase">
          Landing Page
        </div>
        <h1 className="mt-6 text-4xl font-semibold tracking-tight text-foreground">TODO</h1>
        <p className="mt-4 max-w-2xl text-sm leading-6 text-muted-foreground">
          这里先放一个简单的首页占位，后续可以再补充正式内容与导航信息。
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Link href="/dashboard">
            进入控制台
          </Link>
          <Link href="/dashboard/login">
            登录后台
          </Link>
        </div>
      </div>
    </main>
  )
}
