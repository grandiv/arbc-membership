"use client";

import { useEffect, useState, useCallback } from "react";
import { Plus, RefreshCw } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, CAMPAIGN_MENUS, type Member } from "@/lib/api";

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

export default function DashboardPage() {
  const [members, setMembers] = useState<Member[]>([]);
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [menuStats, setMenuStats] = useState<Record<string, number>>({});
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [form, setForm] = useState({ name: "200 Kopi Gratis", limit: "200" });
  const [creating, setCreating] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [m, c, ms] = await Promise.all([api.listMembers(), api.listCampaigns(), api.menuStats()]);
      setMembers(m.data ?? []);
      setCampaigns((c.data as Campaign[]) ?? []);
      setMenuStats(ms.data ?? {});
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal memuat data.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  // The active free-cup campaign (newest active one tagged kind=campaign).
  const active = campaigns
    .filter((c) => c.is_active && c.metadata?.kind === "campaign")
    .sort((a, b) => b.id.localeCompare(a.id))[0];
  const cap = active?.usage?.max_total ?? 0;
  const used = active?.usage_count ?? 0;
  const remaining = cap > 0 ? Math.max(0, cap - used) : 0;

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
          <button className="btn btn--ghost btn--sm" onClick={load} disabled={loading} aria-label="Muat ulang">
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
              <h3 style={{ marginBottom: "1rem" }}>Pendaftar</h3>
              {members.length === 0 ? (
                <p className="muted">Belum ada pendaftar.</p>
              ) : (
                <div className="table-wrap">
                  <table className="table table--cards">
                    <thead><tr><th>Nama</th><th>Nomor HP</th><th>Domisili</th><th>Umur</th><th>Menu</th><th>Status</th></tr></thead>
                    <tbody>
                      {members.map((m) => (
                        <tr key={m.id}>
                          <td data-label="Nama" style={{ fontWeight: 600 }}>{m.name}</td>
                          <td data-label="Nomor HP">{m.phone}</td>
                          <td data-label="Domisili">{m.address ?? "—"}</td>
                          <td data-label="Umur">{umurFromDob(m.date_of_birth)}</td>
                          <td data-label="Menu">{m.menu || "—"}</td>
                          <td data-label="Status">
                            <span className={`pill ${m.order_count > 0 ? "pill--ok" : "pill--warn"}`}>
                              {m.order_count > 0 ? "✓ Klaim" : "Belum"}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
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
