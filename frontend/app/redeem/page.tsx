"use client";

import { useState } from "react";
import { Search, Ticket, CheckCircle2, XCircle } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, type Member } from "@/lib/api";

export default function RedeemPage() {
  // Lookup
  const [phone, setPhone] = useState("");
  const [member, setMember] = useState<Member | null | undefined>(undefined);
  const [looking, setLooking] = useState(false);
  // Redeem
  const [code, setCode] = useState("");
  const [redeeming, setRedeeming] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  async function lookup(e: React.FormEvent) {
    e.preventDefault();
    setMember(undefined);
    setMsg(null);
    setLooking(true);
    try {
      const res = await api.lookup(phone.trim());
      setMember(res.member);
    } catch {
      setMember(null);
    } finally {
      setLooking(false);
    }
  }

  async function redeem(e: React.FormEvent) {
    e.preventDefault();
    setMsg(null);
    setRedeeming(true);
    try {
      const res = await api.redeem({ code: code.trim().toUpperCase(), phone: phone.trim() || undefined });
      setMsg({ ok: true, text: `Voucher terpakai. Diskon Rp${res.discountAmount.toLocaleString("id-ID")}.` });
      setCode("");
    } catch (err) {
      setMsg({ ok: false, text: err instanceof ApiError ? err.message : "Voucher tidak valid." });
    } finally {
      setRedeeming(false);
    }
  }

  return (
    <>
      <Header />
      <main className="wrap" style={{ maxWidth: 560, paddingTop: "2rem", paddingBottom: "4rem" }}>
        <h1 style={{ fontSize: "clamp(1.7rem, 5vw, 2.2rem)", marginBottom: "0.4rem" }}>Tukar Voucher</h1>
        <p className="muted" style={{ marginBottom: "1.75rem" }}>Untuk barista — cari member lalu pakai kode voucher.</p>

        {/* Lookup by phone */}
        <form className="card" style={{ padding: "1.4rem 1.5rem", marginBottom: "1.25rem" }} onSubmit={lookup}>
          <div className="field" style={{ marginBottom: "0.75rem" }}>
            <label htmlFor="lphone">Cari member (nomor HP)</label>
            <input id="lphone" className="input" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="08xxxxxxxxxx" inputMode="tel" />
          </div>
          <button className="btn btn--ghost btn--block" disabled={looking || !phone.trim()} type="submit">
            <Search size={16} /> {looking ? "Mencari…" : "Cari"}
          </button>

          {member && (
            <div className="notice notice--ok" style={{ marginTop: "1rem" }}>
              <strong>{member.name}</strong> · {member.phone}
              <br />
              <span style={{ fontSize: "0.85rem" }}>
                {member.order_count}× kunjungan · total Rp{member.total_spend.toLocaleString("id-ID")}
              </span>
            </div>
          )}
          {member === null && (
            <div className="notice notice--err" style={{ marginTop: "1rem" }}>
              Nomor ini belum terdaftar. Arahkan untuk daftar dulu di /join.
            </div>
          )}
        </form>

        {/* Redeem a code */}
        <form className="card" style={{ padding: "1.4rem 1.5rem" }} onSubmit={redeem}>
          <div className="field" style={{ marginBottom: "0.75rem" }}>
            <label htmlFor="code">Kode voucher</label>
            <input id="code" className="input" value={code} onChange={(e) => setCode(e.target.value)} placeholder="ARBC-XXXXXX" style={{ textTransform: "uppercase", letterSpacing: "0.05em" }} />
          </div>
          <button className="btn btn--primary btn--block" disabled={redeeming || !code.trim()} type="submit">
            <Ticket size={16} /> {redeeming ? "Memproses…" : "Pakai voucher"}
          </button>

          {msg && (
            <div className={`notice ${msg.ok ? "notice--ok" : "notice--err"}`} style={{ marginTop: "1rem", display: "flex", gap: "0.5rem", alignItems: "center" }}>
              {msg.ok ? <CheckCircle2 size={18} /> : <XCircle size={18} />} {msg.text}
            </div>
          )}
        </form>
      </main>
    </>
  );
}
