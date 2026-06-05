import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Static HTML export → served by nginx (same model as the TM/Ijul FEs).
  // All pages are client-rendered, so no server runtime is needed.
  output: "export",
  images: { unoptimized: true },
};

export default nextConfig;
