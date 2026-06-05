"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const links = [
  { href: "/join", label: "Daftar" },
  { href: "/redeem", label: "Tukar" },
  { href: "/admin", label: "Admin" },
];

export default function Header() {
  const path = usePathname();
  return (
    <header style={{ padding: "1.25rem 0" }}>
      <div className="wrap" style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: "0.75rem", flexWrap: "wrap" }}>
        <Link href="/" style={{ fontFamily: "var(--font-display)", fontWeight: 800, fontSize: "1.15rem", letterSpacing: "-0.02em" }}>
          tana arabica <span className="muted" style={{ fontWeight: 600 }}>· member</span>
        </Link>
        <nav className="nav">
          {links.map((l) => (
            <Link key={l.href} href={l.href} className={path?.startsWith(l.href) ? "active" : ""}>
              {l.label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
  );
}
