"use client";

import { useState } from "react";
import { CheckCircle2, XCircle, RotateCcw, Coffee } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, CAMPAIGN_MENUS, type CampaignMenu, type ClaimResult } from "@/lib/api";

export default function ClaimPage() {
  const [form, setForm] = useState({ name: "", phone: "", domisili: "", umur: "" });
  const [menu, setMenu] = useState<CampaignMenu | "">("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ClaimResult | null>(null);

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (!menu) {
      setError("Pilih dulu menu kopinya.");
      return;
    }
    setLoading(true);
    try {
      const res = await api.claim({
        name: form.name.trim(),
        phone: form.phone.trim(),
        domisili: form.domisili.trim() || undefined,
        umur: form.umur ? Number(form.umur) : undefined,
        menu,
      });
      setResult(res);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal. Coba lagi.");
    } finally {
      setLoading(false);
    }
  }

  function reset() {
    setForm({ name: "", phone: "", domisili: "", umur: "" });
    setMenu("");
    setResult(null);
    setError(null);
  }

  const reasonText: Record<string, string> = {
    already_claimed: "Nomor ini sudah pernah klaim kopi gratis. Satu nomor satu cup ya.",
    exhausted: "Kuota kopi gratis sudah habis. Datanya tetap tersimpan 🙌",
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
                <p style={{ marginTop: "0.4rem" }}>
                  Berikan <strong>{result.menu ?? "1 cup"}</strong> ke <strong>{result.member.name}</strong>.
                </p>
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
            <div style={{ marginBottom: "1.5rem", textAlign: "center" }}>
              <img src="/logo-mark.png" alt="" style={{ height: 76, width: "auto", margin: "0 auto 0.75rem" }} />
              <h1 style={{ fontSize: "clamp(1.8rem, 5vw, 2.4rem)", margin: "0 0 0.5rem" }}>
                Klaim <span className="outlined">Kopi Gratis</span>
              </h1>
              <p className="muted" style={{ maxWidth: 380, margin: "0 auto" }}>
                Isikan data pelanggan untuk klaim 1 cup gratis.
              </p>
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
              <div style={{ display: "grid", gridTemplateColumns: "1.6fr 1fr", gap: "1rem" }}>
                <div className="field" style={{ marginBottom: 0 }}>
                  <label htmlFor="domisili">Domisili *</label>
                  <input id="domisili" className="input" required value={form.domisili} onChange={set("domisili")} placeholder="cth. Yogyakarta" autoComplete="off" />
                </div>
                <div className="field" style={{ marginBottom: 0 }}>
                  <label htmlFor="umur">Umur *</label>
                  <input id="umur" className="input" required type="number" min={5} max={100} value={form.umur} onChange={set("umur")} placeholder="cth. 25" inputMode="numeric" />
                </div>
              </div>

              <div className="field" style={{ marginTop: "1.25rem", marginBottom: 0 }}>
                <label>Pilih menu kopi *</label>
                <div className="menu-choice">
                  {CAMPAIGN_MENUS.map((m) => (
                    <button
                      key={m}
                      type="button"
                      className={`menu-opt ${menu === m ? "menu-opt--on" : ""}`}
                      aria-pressed={menu === m}
                      onClick={() => setMenu(m)}
                    >
                      <Coffee size={18} />
                      <span>{m}</span>
                    </button>
                  ))}
                </div>
              </div>

              {error && <div className="notice notice--err" style={{ margin: "1.25rem 0 0" }}>{error}</div>}

              <button className="btn btn--primary btn--block" style={{ marginTop: "1.5rem" }} disabled={loading || !menu} type="submit">
                {loading ? "Memproses…" : "Klaim 1 Kopi Gratis"}
              </button>
            </form>
          </>
        )}
      </main>
    </>
  );
}
