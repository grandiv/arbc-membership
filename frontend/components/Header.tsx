"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

// Staff-only nav (the consumer never operates the app — staff do all inputs).
const links = [
  { href: "/", label: "Klaim" },
  { href: "/produksi", label: "Produksi" },
  { href: "/admin", label: "Dashboard" },
];

export default function Header() {
  const path = usePathname();
  return (
    <header style={{ padding: "1.1rem 0" }}>
      <div className="wrap" style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: "0.75rem", flexWrap: "wrap" }}>
        <Link href="/" aria-label="dan Arabica" style={{ display: "inline-flex", alignItems: "center", gap: "0.55rem" }}>
          {/* Logo mark + wordmark, terracotta on the cream ground. */}
          <img src="/logo-mark.png" alt="" style={{ height: 34, width: "auto" }} />
          <img src="/logo-type.png" alt="dan Arabica" style={{ height: 22, width: "auto" }} />
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
