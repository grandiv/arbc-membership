"use client";

// HIDDEN membership screen — see ./README.md. Consumer self-signup, decoupled
// from the campaign (registration is its own concern; the free-cup perk is a
// separate GET /api/campaign read).
import { useState } from "react";
import { Coffee, CheckCircle2 } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, type RegisterResult, type Campaign } from "@/lib/api";

export default function JoinScreen() {
  const [form, setForm] = useState({ name: "", phone: "", email: "", ig_handle: "", dob: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<RegisterResult | null>(null);
  const [campaign, setCampaign] = useState<Campaign | null>(null);

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await api.register({
        name: form.name.trim(),
        phone: form.phone.trim(),
        email: form.email.trim() || undefined,
        ig_handle: form.ig_handle.trim() || undefined,
        dob: form.dob || undefined,
      });
      setResult(res);
      try {
        setCampaign(await api.campaign());
      } catch {
        setCampaign(null);
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal mendaftar. Coba lagi.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <Header />
      <main className="wrap" style={{ maxWidth: 540, paddingTop: "2rem", paddingBottom: "4rem" }}>
        {result ? (
          <div className="card" style={{ padding: "2rem 1.75rem", textAlign: "center" }}>
            <CheckCircle2 size={44} color="var(--sage)" style={{ margin: "0 auto" }} />
            <h2 style={{ marginTop: "0.75rem" }}>Selamat datang, {result.member.name}! 🎉</h2>
            <p className="muted" style={{ marginTop: "0.25rem" }}>Kamu sekarang member Tanarabica.</p>
            {campaign?.active && (
              <div style={{ marginTop: "1.5rem", paddingTop: "1.25rem", borderTop: "1px solid var(--line)" }}>
                <div className="code">☕ 1 KOPI GRATIS</div>
                <p style={{ marginTop: "1.25rem" }}>
                  Tunjukkan <strong>nomor HP-mu</strong> ke barista di booth untuk klaim.
                </p>
                {typeof campaign.remaining === "number" && campaign.remaining >= 0 && (
                  <p className="muted" style={{ fontSize: "0.85rem", marginTop: "0.5rem" }}>Sisa kuota: {campaign.remaining} cup</p>
                )}
              </div>
            )}
          </div>
        ) : (
          <>
            <div style={{ marginBottom: "1.5rem" }}>
              <Coffee size={30} color="var(--primary)" />
              <h1 style={{ fontSize: "clamp(1.8rem, 5vw, 2.4rem)", margin: "0.5rem 0" }}>
                Jadi <span className="outlined">member</span> Tanarabica.
              </h1>
              <p className="muted">Cukup sekali isi — nomor HP-mu jadi kartu member-mu.</p>
            </div>
            <form className="card" style={{ padding: "1.75rem 1.6rem" }} onSubmit={submit}>
              <div className="field"><label htmlFor="name">Nama *</label>
                <input id="name" className="input" required value={form.name} onChange={set("name")} placeholder="Nama kamu" /></div>
              <div className="field"><label htmlFor="phone">Nomor HP / WhatsApp *</label>
                <input id="phone" className="input" required value={form.phone} onChange={set("phone")} placeholder="08xxxxxxxxxx" inputMode="tel" /></div>
              <div className="field"><label htmlFor="email">Email</label>
                <input id="email" className="input" type="email" value={form.email} onChange={set("email")} placeholder="email@kamu.com (opsional)" /></div>
              <div className="field"><label htmlFor="dob">Tanggal Lahir 🎂</label>
                <input id="dob" className="input" type="date" value={form.dob} onChange={set("dob")} max={new Date().toISOString().slice(0, 10)} /></div>
              <div className="field"><label htmlFor="ig">Instagram</label>
                <input id="ig" className="input" value={form.ig_handle} onChange={set("ig_handle")} placeholder="@username (opsional)" /></div>
              {error && <div className="notice notice--err" style={{ marginBottom: "1rem" }}>{error}</div>}
              <button className="btn btn--primary btn--block" disabled={loading} type="submit">
                {loading ? "Mendaftar…" : "Daftar jadi member"}
              </button>
            </form>
          </>
        )}
      </main>
    </>
  );
}
