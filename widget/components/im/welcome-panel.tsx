"use client";

type WelcomePanelProps = {
  title: string;
  welcomeText?: string;
};

export function WelcomePanel({ title, welcomeText }: WelcomePanelProps) {
  return (
    <section className="cs-agent-fade-up relative overflow-hidden rounded-3xl border border-white/80 bg-[linear-gradient(135deg,rgba(255,255,255,0.96),rgba(244,248,252,0.92))] p-4 shadow-[0_14px_30px_rgba(15,23,42,0.06)] backdrop-blur">
      <div className="absolute inset-x-6 top-0 h-px bg-linear-to-r from-transparent via-sky-200 to-transparent" />
      <h1 className="text-[20px] leading-7 font-semibold text-slate-950">
        {title}
      </h1>
      <p className="mt-2 text-sm leading-6 text-slate-600">
        {welcomeText ||
          "你好，这里是在线客服。你可以直接发送消息，我们会为你匹配当前可用的服务人员。"}
      </p>
      <div className="mt-4 rounded-2xl border border-slate-200/80 bg-white/80 px-3 py-2 text-[12px] leading-5 text-slate-500">
        工作时间内通常几分钟内回复，关闭窗口后会话记录会自动保留。
      </div>
    </section>
  );
}
