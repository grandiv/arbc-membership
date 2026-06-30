"use client";

import { useEffect, useMemo, useRef, useState, useCallback } from "react";
import { Plus, RefreshCw, Download } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, CAMPAIGN_MENUS, type Member } from "@/lib/api";
import type { SheetData } from "write-excel-file/browser";

const PAGE_SIZE = 15;

type Campaign = {
  id: string;
  code: string;
  is_active: boolean;
  usage_count: number;
  usage?: { max_total?: number };
  metadata?: { campaign?: string; label?: string; kind?: string };
};

// Umur disimpan sebagai date_of_birth (1 Jan tahun lahir); tampilkan kembali sebagai angka.
function umurFromDob(dob?: string | null): string {
  if (!dob || dob.length < 4) return "—";
  const year = parseInt(dob.slice(0, 4), 10);
  if (!year) return "—";
  return String(new Date().getFullYear() - year);
}

// Live production status → label + pill class. Falls back to claim status.
function statusBadge(m: Member): { label: string; cls: string } {
  switch (m.queue_status) {
    case "waiting": return { label: "Sedang dibuat", cls: "pill--making" };
    case "ready":   return { label: "Siap dipanggil", cls: "pill--ready" };
    case "done":    return { label: "Selesai", cls: "pill--done" };
    default:        return m.order_count > 0
      ? { label: "✓ Klaim", cls: "pill--ok" }
      : { label: "Belum", cls: "pill--warn" };
  }
}

export default function DashboardPage() {
  const [members, setMembers] = useState<Member[]>([]);
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [menuStats, setMenuStats] = useState<Record<string, number>>({});
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [form, setForm] = useState({ name: "200 Kopi Gratis", limit: "200" });
  const [creating, setCreating] = useState(false);
  const [page, setPage] = useState(1);
  const [exporting, setExporting] = useState(false);

  // load(silent) — manual/initial loads show the spinner; the 3s background
  // poll refreshes quietly so the dashboard stays live without flicker.
  const load = useCallback(async (silent = false) => {
    if (!silent) setLoading(true);
    try {
      const [m, c, ms] = await Promise.all([api.listMembers(), api.listCampaigns(), api.menuStats()]);
      setMembers(m.data ?? []);
      setCampaigns((c.data as Campaign[]) ?? []);
      setMenuStats(ms.data ?? {});
      setError(null);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal memuat data.");
    } finally {
      if (!silent) setLoading(false);
    }
  }, []);

  // Auto-refresh every 3s, paused while the tab is hidden (don't poll Atlas when
  // nobody's looking); refresh immediately when the tab becomes visible again.
  const POLL_MS = 3000;
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    let alive = true;
    const tick = async () => {
      if (!document.hidden) await load(true);
      if (alive) timer.current = setTimeout(tick, POLL_MS);
    };
    load(); // initial (with spinner)
    timer.current = setTimeout(tick, POLL_MS);

    const onVisible = () => { if (!document.hidden) load(true); };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      alive = false;
      if (timer.current) clearTimeout(timer.current);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, [load]);

  // The active free-cup campaign (newest active one tagged kind=campaign).
  const active = campaigns
    .filter((c) => c.is_active && c.metadata?.kind === "campaign")
    .sort((a, b) => b.id.localeCompare(a.id))[0];
  const cap = active?.usage?.max_total ?? 0;
  const used = active?.usage_count ?? 0;
  const remaining = cap > 0 ? Math.max(0, cap - used) : 0;

  // Newest claim first (by queue number); members without a number sink to the
  // bottom. Sorted once per data change, then sliced for the current page.
  const sorted = useMemo(
    () => [...members].sort((a, b) => (b.queue_number ?? 0) - (a.queue_number ?? 0)),
    [members]
  );
  const totalPages = Math.max(1, Math.ceil(sorted.length / PAGE_SIZE));
  const safePage = Math.min(page, totalPages);
  const pageStart = (safePage - 1) * PAGE_SIZE;
  const pageRows = sorted.slice(pageStart, pageStart + PAGE_SIZE);
  // New claims arrive on page 1 (newest-first), so paging stays put; just clamp.
  useEffect(() => {
    if (page > totalPages) setPage(totalPages);
  }, [page, totalPages]);

  // Export ALL pendaftar to a real .xlsx. Phone is a text column so Excel keeps
  // the leading 0 (and doesn't switch to scientific notation). Lazy-loaded.
  async function exportExcel() {
    setExporting(true);
    try {
      const writeXlsxFile = (await import("write-excel-file/browser")).default;
      const rows = [...members].sort((a, b) => (a.queue_number ?? 0) - (b.queue_number ?? 0));
      const HEADER = ["Antrian", "Nama", "Nomor HP", "Domisili", "Umur", "Menu", "Status"].map((value) => ({
        value, fontWeight: "bold" as const,
      }));
      const data: SheetData = [
        HEADER,
        ...rows.map((m) => [
          { type: Number, value: m.queue_number || undefined },
          { type: String, value: m.name || "" },
          { type: String, value: m.phone || "" }, // text → keeps leading 0
          { type: String, value: m.address || "" },
          { type: String, value: umurFromDob(m.date_of_birth) },
          { type: String, value: m.menu || "" },
          { type: String, value: statusBadge(m).label.replace("✓ ", "") },
        ]),
      ];
      const today = new Date().toISOString().slice(0, 10);
      // Browser build returns { toBlob, toFile } — call toFile to download.
      await writeXlsxFile(data, {
        columns: [{ width: 10 }, { width: 22 }, { width: 18 }, { width: 18 }, { width: 8 }, { width: 18 }, { width: 16 }],
      }).toFile(`dan-arabica-pendaftar-${today}.xlsx`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal mengekspor.");
    } finally {
      setExporting(false);
    }
  }

  async function createCampaign(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError(null);
    try {
      const code = "KOPI" + Date.now().toString().slice(-6);
      await api.createCampaign({
        code,
        name: form.name.trim(),
        limit: parseInt(form.limit, 10) || 200,
        discount_type: "free",
      });
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal membuat kampanye.");
    } finally {
      setCreating(false);
    }
  }

  return (
    <>
      <Header />
      <main className="wrap" style={{ paddingTop: "2rem", paddingBottom: "4rem" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: "0.75rem", flexWrap: "wrap", marginBottom: "1.5rem" }}>
          <h1 style={{ fontSize: "clamp(1.7rem, 5vw, 2.2rem)" }}>Dashboard Kampanye</h1>
          <button className="btn btn--ghost btn--sm" onClick={() => load()} disabled={loading} aria-label="Muat ulang">
            <RefreshCw size={15} /> <span className="btn__label">Muat ulang</span>
          </button>
        </div>

        {error && <div className="notice notice--err" style={{ marginBottom: "1.5rem" }}>{error}</div>}

        {active ? (
          <>
            {/* Campaign progress */}
            <div className="grid-2" style={{ marginBottom: "1.25rem" }}>
              <div className="stat"><div className="stat__n">{used}</div><div className="stat__l">Kopi Gratis Terpakai</div></div>
              <div className="stat"><div className="stat__n">{remaining}</div><div className="stat__l">Sisa Kuota</div></div>
              <div className="stat"><div className="stat__n">{cap}</div><div className="stat__l">Total Kuota</div></div>
              <div className="stat"><div className="stat__n">{members.length}</div><div className="stat__l">Total Pendaftar</div></div>
            </div>

            {/* Menu breakdown — which free drink is more popular */}
            <div className="grid-2" style={{ marginBottom: "2rem" }}>
              {CAMPAIGN_MENUS.map((m) => (
                <div className="stat" key={m}>
                  <div className="stat__n">{menuStats[m] ?? 0}</div>
                  <div className="stat__l">{m}</div>
                </div>
              ))}
            </div>

            {/* Pendaftar — the data collected (Nama + HP), with claim status */}
            <section className="card" style={{ padding: "1.5rem 1.6rem" }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: "0.75rem", flexWrap: "wrap", marginBottom: "1rem" }}>
                <h3 style={{ margin: 0 }}>Pendaftar <span className="muted" style={{ fontWeight: 400 }}>({members.length})</span></h3>
                <button className="btn btn--ghost btn--sm" onClick={exportExcel} disabled={exporting || members.length === 0}>
                  <Download size={15} /> <span className="btn__label">{exporting ? "Mengekspor…" : "Ekspor ke Excel"}</span>
                </button>
              </div>
              {members.length === 0 ? (
                <p className="muted">Belum ada pendaftar.</p>
              ) : (
                <>
                  <div className="table-wrap">
                    <table className="table table--cards">
                      <thead><tr><th>Antrian</th><th>Nama</th><th>Nomor HP</th><th>Domisili</th><th>Umur</th><th>Menu</th><th>Status</th></tr></thead>
                      <tbody>
                        {pageRows.map((m) => {
                          const s = statusBadge(m);
                          return (
                          <tr key={m.id}>
                            <td data-label="Antrian"><span className="qnum">{m.queue_number ? `#${m.queue_number}` : "—"}</span></td>
                            <td data-label="Nama" style={{ fontWeight: 600 }}>{m.name}</td>
                            <td data-label="Nomor HP">{m.phone}</td>
                            <td data-label="Domisili">{m.address ?? "—"}</td>
                            <td data-label="Umur">{umurFromDob(m.date_of_birth)}</td>
                            <td data-label="Menu">{m.menu || "—"}</td>
                            <td data-label="Status">
                              <span className={`pill ${s.cls}`}>{s.label}</span>
                            </td>
                          </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                  {totalPages > 1 && (
                    <div className="pager">
                      <span className="muted">
                        Menampilkan {pageStart + 1}–{Math.min(pageStart + PAGE_SIZE, sorted.length)} dari {sorted.length}
                      </span>
                      <div className="pager__btns">
                        <button className="btn btn--ghost btn--sm" disabled={safePage <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))}>← Sebelumnya</button>
                        <span className="pager__page">Hal {safePage} / {totalPages}</span>
                        <button className="btn btn--ghost btn--sm" disabled={safePage >= totalPages} onClick={() => setPage((p) => Math.min(totalPages, p + 1))}>Berikutnya →</button>
                      </div>
                    </div>
                  )}
                </>
              )}
            </section>
          </>
        ) : (
          // No active campaign → one-tap setup.
          <section className="card" style={{ padding: "1.6rem" }}>
            <h3 style={{ marginBottom: "0.5rem" }}>Mulai kampanye</h3>
            <p className="muted" style={{ marginBottom: "1.25rem" }}>Belum ada kampanye aktif. Buat kampanye kopi gratis untuk mulai.</p>
            <form onSubmit={createCampaign} style={{ display: "grid", gap: "1rem", gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))", alignItems: "end" }}>
              <div className="field" style={{ margin: 0 }}>
                <label htmlFor="cname">Nama kampanye</label>
                <input id="cname" className="input" required value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} />
              </div>
              <div className="field" style={{ margin: 0 }}>
                <label htmlFor="climit">Kuota cup</label>
                <input id="climit" className="input" type="number" min={1} value={form.limit} onChange={(e) => setForm((f) => ({ ...f, limit: e.target.value }))} />
              </div>
              <button className="btn btn--primary" disabled={creating} type="submit">
                <Plus size={16} /> {creating ? "Membuat…" : "Mulai"}
              </button>
            </form>
          </section>
        )}
      </main>
    </>
  );
}
