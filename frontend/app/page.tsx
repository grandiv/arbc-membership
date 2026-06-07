"use client";

import { useEffect, useState } from "react";
import { Coffee, CheckCircle2, XCircle, RotateCcw } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, type ClaimResult, type Campaign } from "@/lib/api";

export default function ClaimPage() {
  const [form, setForm] = useState({ name: "", phone: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ClaimResult | null>(null);
  const [campaign, setCampaign] = useState<Campaign | null>(null);

  const loadCampaign = () => api.campaign().then(setCampaign).catch(() => setCampaign(null));
  useEffect(() => {
    loadCampaign();
  }, []);

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await api.claim({ name: form.name.trim(), phone: form.phone.trim() });
      setResult(res);
      loadCampaign();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal. Coba lagi.");
    } finally {
      setLoading(false);
    }
  }

  function reset() {
    setForm({ name: "", phone: "" });
    setResult(null);
    setError(null);
  }

  const reasonText: Record<string, string> = {
    already_claimed: "Nomor ini sudah pernah klaim kopi gratis. Satu nomor satu cup ya.",
    exhausted: "Kuota 200 kopi gratis sudah habis. Datanya tetap tersimpan 🙌",
    no_campaign: "Belum ada kampanye aktif. Buat dulu di Dashboard.",
    ineligible: "Belum bisa diklaim saat ini.",
  };

  return (
    <>
      <Header />
      <main className="wrap" style={{ maxWidth: 520, paddingTop: "2rem", paddingBottom: "4rem" }}>
        {result ? (
          <div className="card" style={{ padding: "2rem 1.75rem", textAlign: "center" }}>
            {result.claimed ? (
              <>
                <CheckCircle2 size={48} color="var(--sage)" style={{ margin: "0 auto" }} />
                <h2 style={{ marginTop: "0.75rem" }}>Kopi gratis! ☕</h2>
                <p style={{ marginTop: "0.4rem" }}>Berikan 1 cup ke <strong>{result.member.name}</strong>.</p>
                {typeof result.remaining === "number" && result.remaining >= 0 && (
                  <p className="muted" style={{ marginTop: "0.6rem" }}>Sisa kuota: {result.remaining} / 200</p>
                )}
              </>
            ) : (
              <>
                <XCircle size={48} color="var(--danger)" style={{ margin: "0 auto" }} />
                <h2 style={{ marginTop: "0.75rem" }}>Belum bisa</h2>
                <p className="muted" style={{ marginTop: "0.5rem" }}>{reasonText[result.reason ?? "ineligible"]}</p>
              </>
            )}
            <button className="btn btn--primary btn--block" style={{ marginTop: "1.5rem" }} onClick={reset}>
              <RotateCcw size={16} /> Pelanggan berikutnya
            </button>
          </div>
        ) : (
          <>
            <div style={{ marginBottom: "1.5rem" }}>
              <Coffee size={30} color="var(--caramel)" />
              <h1 style={{ fontSize: "clamp(1.8rem, 5vw, 2.4rem)", margin: "0.5rem 0" }}>
                Klaim <span className="outlined">200 Kopi Gratis</span>
              </h1>
              <p className="muted">Isikan data pelanggan untuk klaim 1 cup gratis. Satu nomor satu cup.</p>
              {campaign?.active && typeof campaign.remaining === "number" && campaign.remaining >= 0 && (
                <p style={{ marginTop: "0.5rem", fontWeight: 600, color: "var(--sage)" }}>
                  ☕ Sisa kuota: {campaign.remaining} / 200
                </p>
              )}
            </div>

            <form className="card" style={{ padding: "1.75rem 1.6rem" }} onSubmit={submit}>
              <div className="field">
                <label htmlFor="name">Nama *</label>
                <input id="name" className="input" required value={form.name} onChange={set("name")} placeholder="Nama pelanggan" autoComplete="off" />
              </div>
              <div className="field">
                <label htmlFor="phone">Nomor HP / WhatsApp *</label>
                <input id="phone" className="input" required value={form.phone} onChange={set("phone")} placeholder="08xxxxxxxxxx" inputMode="tel" autoComplete="off" />
              </div>

              {error && <div className="notice notice--err" style={{ marginBottom: "1rem" }}>{error}</div>}

              <button className="btn btn--primary btn--block" disabled={loading} type="submit">
                {loading ? "Memproses…" : "Klaim 1 Kopi Gratis"}
              </button>
            </form>
          </>
        )}
      </main>
    </>
  );
}
