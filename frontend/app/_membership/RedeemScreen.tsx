"use client";

// HIDDEN membership screen — see ./README.md. Barista flow: look up a member by
// phone, redeem the active campaign by phone (or an explicit code); price-aware.
import { useState } from "react";
import { Search, Coffee, CheckCircle2, XCircle, Ticket } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, type Member, type FreeCup } from "@/lib/api";

export default function RedeemScreen() {
  const [phone, setPhone] = useState("");
  const [member, setMember] = useState<Member | null | undefined>(undefined);
  const [freeCup, setFreeCup] = useState<FreeCup | null>(null);
  const [amount, setAmount] = useState("");
  const [code, setCode] = useState("");
  const [looking, setLooking] = useState(false);
  const [redeeming, setRedeeming] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  async function lookup(e: React.FormEvent) {
    e.preventDefault();
    setMember(undefined);
    setFreeCup(null);
    setMsg(null);
    setLooking(true);
    try {
      const res = await api.lookup(phone.trim());
      setMember(res.member);
      setFreeCup(res.freeCup);
    } catch {
      setMember(null);
    } finally {
      setLooking(false);
    }
  }

  async function redeem(opts: { code?: string }) {
    setMsg(null);
    setRedeeming(true);
    try {
      const res = await api.redeem({
        phone: phone.trim() || undefined,
        code: opts.code,
        name: member?.name,
        amount: amount ? Number(amount) : undefined,
      });
      setMsg({ ok: true, text: `☕ Kopi gratis! Potongan Rp${res.discountAmount.toLocaleString("id-ID")} dari Rp${res.orderAmount.toLocaleString("id-ID")}.` });
      setCode("");
      if (phone.trim()) {
        const r = await api.lookup(phone.trim());
        setFreeCup(r.freeCup);
      }
    } catch (err) {
      setMsg({ ok: false, text: err instanceof ApiError ? err.message : "Gagal menukar." });
    } finally {
      setRedeeming(false);
    }
  }

  return (
    <>
      <Header />
      <main className="wrap" style={{ maxWidth: 560, paddingTop: "2rem", paddingBottom: "4rem" }}>
        <h1 style={{ fontSize: "clamp(1.7rem, 5vw, 2.2rem)", marginBottom: "0.4rem" }}>Tukar Kopi Gratis</h1>
        <p className="muted" style={{ marginBottom: "1.75rem" }}>Untuk barista — masukkan nomor HP member.</p>

        <form className="card" style={{ padding: "1.4rem 1.5rem", marginBottom: "1.25rem" }} onSubmit={lookup}>
          <div className="field" style={{ marginBottom: "0.75rem" }}>
            <label htmlFor="lphone">Nomor HP member</label>
            <input id="lphone" className="input" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="08xxxxxxxxxx" inputMode="tel" />
          </div>
          <button className="btn btn--ghost btn--block" disabled={looking || !phone.trim()} type="submit">
            <Search size={16} /> {looking ? "Mencari…" : "Cari member"}
          </button>

          {member && (
            <>
              <div className="notice notice--ok" style={{ marginTop: "1rem" }}>
                <strong>{member.name}</strong> · {member.phone}
                <br /><span style={{ fontSize: "0.85rem" }}>{member.order_count}× kunjungan · total Rp{member.total_spend.toLocaleString("id-ID")}</span>
              </div>
              <div className="field" style={{ marginTop: "1rem", marginBottom: "0.75rem" }}>
                <label htmlFor="amount">Harga pesanan (opsional)</label>
                <input id="amount" className="input" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="cth. 18000" inputMode="numeric" />
              </div>
              {freeCup?.eligible ? (
                <button className="btn btn--primary btn--block" disabled={redeeming} onClick={() => redeem({})} type="button">
                  <Coffee size={16} /> {redeeming ? "Memproses…" : "Tukar Kopi Gratis"}
                </button>
              ) : (
                <div className="notice notice--warn" style={{ marginTop: "0.25rem" }}>Member ini sudah klaim / tidak ada kampanye aktif.</div>
              )}
            </>
          )}
          {member === null && <div className="notice notice--err" style={{ marginTop: "1rem" }}>Nomor ini belum terdaftar.</div>}
        </form>

        <details className="card" style={{ padding: "1.1rem 1.5rem" }}>
          <summary style={{ cursor: "pointer", fontFamily: "var(--font-display)", fontWeight: 600 }}>Pakai kode voucher</summary>
          <div className="field" style={{ marginTop: "1rem", marginBottom: "0.75rem" }}>
            <label htmlFor="code">Kode</label>
            <input id="code" className="input" value={code} onChange={(e) => setCode(e.target.value)} placeholder="KODE" style={{ textTransform: "uppercase" }} />
          </div>
          <button className="btn btn--ghost btn--block" disabled={redeeming || !code.trim()} onClick={() => redeem({ code: code.trim().toUpperCase() })} type="button">
            <Ticket size={16} /> {redeeming ? "Memproses…" : "Pakai kode"}
          </button>
        </details>

        {msg && (
          <div className={`notice ${msg.ok ? "notice--ok" : "notice--err"}`} style={{ marginTop: "1.25rem", display: "flex", gap: "0.5rem", alignItems: "center" }}>
            {msg.ok ? <CheckCircle2 size={18} /> : <XCircle size={18} />} {msg.text}
          </div>
        )}
      </main>
    </>
  );
}
