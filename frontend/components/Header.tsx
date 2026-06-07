"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

// Staff-only nav (the consumer never operates the app — staff do all inputs).
const links = [
  { href: "/", label: "Klaim" },
  { href: "/admin", label: "Dashboard" },
];

export default function Header() {
  const path = usePathname();
  return (
    <header style={{ padding: "1.25rem 0" }}>
      <div className="wrap" style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: "0.75rem", flexWrap: "wrap" }}>
        <Link href="/" style={{ fontFamily: "var(--font-display)", fontWeight: 800, fontSize: "1.15rem", letterSpacing: "-0.02em" }}>
          Tanarabica
        </Link>
        <nav className="nav">
          {links.map((l) => (
            <Link key={l.href} href={l.href} className={(l.href === "/" ? path === "/" : path?.startsWith(l.href)) ? "active" : ""}>
              {l.label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
  );
}
