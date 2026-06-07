"use client";

// HIDDEN membership screen — see ./README.md. Full admin: member list (with
// email/visits/spend) + campaign management with configurable discounts.
import { useEffect, useState, useCallback } from "react";
import { Plus, RefreshCw } from "lucide-react";
import Header from "@/components/Header";
import { api, ApiError, type Member } from "@/lib/api";

type Campaign = {
  id: string;
  code: string;
  is_active: boolean;
  usage_count: number;
  usage?: { max_total?: number };
  metadata?: { campaign?: string; label?: string };
};

export default function MembersDashboard() {
  const [members, setMembers] = useState<Member[]>([]);
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [form, setForm] = useState({ code: "", name: "", limit: "200", discount_type: "free", discount_value: "" });
  const [creating, setCreating] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [m, c] = await Promise.all([api.listMembers(), api.listCampaigns()]);
      setMembers(m.data ?? []);
      setCampaigns((c.data as Campaign[]) ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal memuat data.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  async function createCampaign(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError(null);
    try {
      await api.createCampaign({
        code: form.code.trim().toUpperCase(),
        name: form.name.trim(),
        limit: parseInt(form.limit, 10) || 0,
        discount_type: form.discount_type as "free" | "percent" | "fixed",
        discount_value: form.discount_value ? Number(form.discount_value) : undefined,
      });
      setForm({ code: "", name: "", limit: "200", discount_type: "free", discount_value: "" });
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal membuat kampanye.");
    } finally {
      setCreating(false);
    }
  }

  const totalSpend = members.reduce((s, m) => s + m.total_spend, 0);
  const totalVisits = members.reduce((s, m) => s + m.order_count, 0);

  return (
    <>
      <Header />
      <main className="wrap" style={{ paddingTop: "2rem", paddingBottom: "4rem" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: "0.75rem", flexWrap: "wrap", marginBottom: "1.5rem" }}>
          <h1 style={{ fontSize: "clamp(1.7rem, 5vw, 2.2rem)" }}>Dashboard</h1>
          <button className="btn btn--ghost btn--sm" onClick={load} disabled={loading} aria-label="Muat ulang">
            <RefreshCw size={15} /> <span className="btn__label">Muat ulang</span>
          </button>
        </div>
        {error && <div className="notice notice--err" style={{ marginBottom: "1.5rem" }}>{error}</div>}

        <div className="grid-2" style={{ marginBottom: "2rem" }}>
          <div className="stat"><div className="stat__n">{members.length}</div><div className="stat__l">Total Member</div></div>
          <div className="stat"><div className="stat__n">{totalVisits}</div><div className="stat__l">Total Kunjungan</div></div>
          <div className="stat"><div className="stat__n">Rp{totalSpend.toLocaleString("id-ID")}</div><div className="stat__l">Total Belanja</div></div>
          <div className="stat"><div className="stat__n">{campaigns.length}</div><div className="stat__l">Kampanye</div></div>
        </div>

        <section className="card" style={{ padding: "1.5rem 1.6rem", marginBottom: "2rem" }}>
          <h3 style={{ marginBottom: "1rem" }}>Buat kampanye baru</h3>
          <form onSubmit={createCampaign} style={{ display: "grid", gap: "1rem", gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))", alignItems: "end" }}>
            <div className="field" style={{ margin: 0 }}><label htmlFor="ccode">Kode</label>
              <input id="ccode" className="input" required value={form.code} onChange={(e) => setForm((f) => ({ ...f, code: e.target.value }))} placeholder="ARABICA200" style={{ textTransform: "uppercase" }} /></div>
            <div className="field" style={{ margin: 0 }}><label htmlFor="cname">Nama</label>
              <input id="cname" className="input" required value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} placeholder="200 Kopi Gratis" /></div>
            <div className="field" style={{ margin: 0 }}><label htmlFor="climit">Kuota</label>
              <input id="climit" className="input" type="number" min={1} value={form.limit} onChange={(e) => setForm((f) => ({ ...f, limit: e.target.value }))} /></div>
            <div className="field" style={{ margin: 0 }}><label htmlFor="cdtype">Diskon</label>
              <select id="cdtype" className="input" value={form.discount_type} onChange={(e) => setForm((f) => ({ ...f, discount_type: e.target.value }))}>
                <option value="free">Gratis (100%)</option><option value="percent">Persen (%)</option><option value="fixed">Potongan Rp</option>
              </select></div>
            {form.discount_type !== "free" && (
              <div className="field" style={{ margin: 0 }}><label htmlFor="cdval">{form.discount_type === "percent" ? "Nilai (%)" : "Nilai (Rp)"}</label>
                <input id="cdval" className="input" type="number" min={1} value={form.discount_value} onChange={(e) => setForm((f) => ({ ...f, discount_value: e.target.value }))} /></div>
            )}
            <button className="btn btn--primary" disabled={creating} type="submit"><Plus size={16} /> {creating ? "Membuat…" : "Buat"}</button>
          </form>
        </section>

        <section className="card" style={{ padding: "1.5rem 1.6rem" }}>
          <h3 style={{ marginBottom: "1rem" }}>Member</h3>
          {members.length === 0 ? <p className="muted">Belum ada member.</p> : (
            <div className="table-wrap">
              <table className="table table--cards">
                <thead><tr><th>Nama</th><th>HP</th><th>Email</th><th>Lahir</th><th>Kunjungan</th><th>Belanja</th></tr></thead>
                <tbody>
                  {members.map((m) => (
                    <tr key={m.id}>
                      <td data-label="Nama" style={{ fontWeight: 600 }}>{m.name}</td>
                      <td data-label="HP">{m.phone}</td>
                      <td data-label="Email" className="muted">{m.email ?? "—"}</td>
                      <td data-label="Lahir">{m.date_of_birth ?? "—"}</td>
                      <td data-label="Kunjungan">{m.order_count}</td>
                      <td data-label="Belanja">Rp{m.total_spend.toLocaleString("id-ID")}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </main>
    </>
  );
}
