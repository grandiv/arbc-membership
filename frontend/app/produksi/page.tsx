"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Coffee, BellRing, Check, RefreshCw } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, type QueueTicket } from "@/lib/api";

// Live production board for the production house: see incoming free-cup orders,
// mark a drink ready (calls the customer), then clear it. Polls every 3s — no
// WebSocket needed at booth volume. Data is folded from AgregaZcy ticket events.
const POLL_MS = 3000;

export default function ProduksiPage() {
  const [tickets, setTickets] = useState<QueueTicket[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [busy, setBusy] = useState<string | null>(null); // ticketId being mutated
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await api.queue();
      setTickets(res.data ?? []);
      setError(null);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal memuat antrian.");
    } finally {
      setLoaded(true);
    }
  }, []);

  // Self-rescheduling poll (avoids overlap if a request is slow), paused while
  // the tab is hidden; refresh immediately when it becomes visible again.
  useEffect(() => {
    let alive = true;
    const tick = async () => {
      if (!document.hidden) await load();
      if (alive) timer.current = setTimeout(tick, POLL_MS);
    };
    tick();
    const onVisible = () => { if (!document.hidden) load(); };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      alive = false;
      if (timer.current) clearTimeout(timer.current);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, [load]);

  async function act(id: string, fn: (id: string) => Promise<unknown>) {
    setBusy(id);
    try {
      await fn(id);
      await load(); // reflect immediately, don't wait for the next poll
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal memperbarui antrian.");
    } finally {
      setBusy(null);
    }
  }

  const waiting = tickets.filter((t) => t.status === "waiting");
  const ready = tickets.filter((t) => t.status === "ready");

  return (
    <>
      <Header />
      <main className="wrap" style={{ paddingTop: "2rem", paddingBottom: "4rem" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: "0.75rem", flexWrap: "wrap", marginBottom: "1.5rem" }}>
          <h1 style={{ fontSize: "clamp(1.7rem, 5vw, 2.2rem)" }}>Antrian Produksi</h1>
          <button className="btn btn--ghost btn--sm" onClick={load} aria-label="Muat ulang">
            <RefreshCw size={15} /> <span className="btn__label">Muat ulang</span>
          </button>
        </div>

        {error && <div className="notice notice--err" style={{ marginBottom: "1.5rem" }}>{error}</div>}

        {/* Siap dipanggil — highlighted so staff call the name */}
        <section style={{ marginBottom: "2rem" }}>
          <h3 className="board__head"><BellRing size={18} /> Siap dipanggil ({ready.length})</h3>
          {ready.length === 0 ? (
            <p className="muted">Belum ada yang siap.</p>
          ) : (
            <div className="board">
              {ready.map((t) => (
                <div key={t.ticketId} className="qcard qcard--ready">
                  <div className="qcard__top">
                    <span className="qcard__no">#{t.number || "—"}</span>
                    <span className="pill pill--ok">PANGGIL</span>
                  </div>
                  <div className="qcard__name">{t.name}</div>
                  <div className="qcard__menu"><Coffee size={14} /> {t.menu}</div>
                  <button className="btn btn--primary btn--block" disabled={busy === t.ticketId} onClick={() => act(t.ticketId, api.ticketDone)}>
                    <Check size={16} /> {busy === t.ticketId ? "…" : "Selesai"}
                  </button>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Sedang dibuat */}
        <section>
          <h3 className="board__head"><Coffee size={18} /> Sedang dibuat ({waiting.length})</h3>
          {!loaded ? (
            <p className="muted">Memuat…</p>
          ) : waiting.length === 0 ? (
            <p className="muted">Tidak ada antrian.</p>
          ) : (
            <div className="board">
              {waiting.map((t) => (
                <div key={t.ticketId} className="qcard">
                  <div className="qcard__top">
                    <span className="qcard__no">#{t.number || "—"}</span>
                  </div>
                  <div className="qcard__name">{t.name}</div>
                  <div className="qcard__menu"><Coffee size={14} /> {t.menu}</div>
                  <button className="btn btn--ghost btn--block" disabled={busy === t.ticketId} onClick={() => act(t.ticketId, api.ticketReady)}>
                    <BellRing size={16} /> {busy === t.ticketId ? "…" : "Siap — Panggil"}
                  </button>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>
    </>
  );
}
