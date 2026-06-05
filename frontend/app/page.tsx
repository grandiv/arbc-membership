import Link from "next/link";
import { Coffee, QrCode, BarChart3 } from "lucide-react";
import Header from "@/components/Header";

export default function Home() {
  return (
    <>
      <Header />
      <main className="wrap" style={{ paddingTop: "3rem", paddingBottom: "5rem" }}>
        <p style={{ fontFamily: "var(--font-display)", fontWeight: 600, color: "var(--cocoa-2)", letterSpacing: "0.04em", textTransform: "uppercase", fontSize: "0.8rem" }}>
          Ngopi Enak Tiap Hari
        </p>
        <h1 style={{ fontSize: "clamp(2.4rem, 6vw, 4rem)", margin: "0.5rem 0 1rem", maxWidth: 720 }}>
          Jadi <span className="outlined">member</span>,<br />ngopi makin <span className="outlined">untung</span>.
        </h1>
        <p style={{ fontSize: "1.1rem", maxWidth: 560, marginBottom: "2.5rem" }}>
          Daftar sekali, datanya kami simpan untuk kasih kamu promo &amp; kopi gratis.
          Cukup nomor HP — tanpa aplikasi, tanpa ribet.
        </p>

        <div className="grid-2" style={{ maxWidth: 820 }}>
          <Link href="/join" className="card" style={cardStyle}>
            <Coffee size={28} color="var(--caramel)" />
            <h3 style={{ marginTop: "0.8rem" }}>Daftar Member</h3>
            <p className="muted" style={{ marginTop: "0.3rem", fontSize: "0.95rem" }}>
              Isi data, langsung dapat voucher kopi gratis.
            </p>
          </Link>
          <Link href="/redeem" className="card" style={cardStyle}>
            <QrCode size={28} color="var(--caramel)" />
            <h3 style={{ marginTop: "0.8rem" }}>Tukar Voucher</h3>
            <p className="muted" style={{ marginTop: "0.3rem", fontSize: "0.95rem" }}>
              Untuk barista — cari member &amp; pakai voucher.
            </p>
          </Link>
          <Link href="/admin" className="card" style={cardStyle}>
            <BarChart3 size={28} color="var(--caramel)" />
            <h3 style={{ marginTop: "0.8rem" }}>Dashboard</h3>
            <p className="muted" style={{ marginTop: "0.3rem", fontSize: "0.95rem" }}>
              Member, kampanye, &amp; analitik bisnis.
            </p>
          </Link>
        </div>
      </main>
    </>
  );
}

const cardStyle: React.CSSProperties = {
  padding: "1.5rem 1.6rem",
  display: "block",
};
