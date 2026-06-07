// Typed client for the arbc-membership BFF. The FE talks ONLY to the BFF —
// never to a KreaZcy engine directly.

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type Member = {
  id: string;
  customer_id: string;
  phone: string;
  name: string;
  email?: string | null;
  date_of_birth?: string | null;
  order_count: number;
  total_spend: number;
};

export type RegisterResult = {
  member: Member;
};

// Result of the staff claim action (register + redeem the free cup).
export type ClaimResult = {
  claimed: boolean;
  reason?: "no_campaign" | "already_claimed" | "exhausted" | "ineligible";
  member: Member;
  remaining?: number;
};

// Active campaign info.
export type Campaign = {
  active: boolean;
  label?: string;
  remaining?: number; // -1 = unlimited
};

export type FreeCup = {
  eligible: boolean;
  label?: string;
  remaining?: number;
};

export type LookupResult = {
  member: Member | null;
  freeCup: FreeCup;
};

export type RedeemResult = {
  redeemed: boolean;
  discountAmount: number;
  orderAmount: number;
};

export class ApiError extends Error {
  code: string;
  status: number;
  constructor(status: number, code: string, message: string) {
    super(message || code);
    this.status = status;
    this.code = code;
  }
}

async function call<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${BASE}${path}`, {
      ...init,
      headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    });
  } catch {
    throw new ApiError(0, "NETWORK", "Tidak bisa terhubung ke server.");
  }
  const text = await res.text();
  const body = text ? JSON.parse(text) : {};
  if (!res.ok) {
    throw new ApiError(res.status, body.code ?? "ERROR", body.message ?? "Terjadi kesalahan.");
  }
  return body as T;
}

export const api = {
  // The campaign's single staff action: capture name+phone AND claim the cup.
  claim: (input: { name: string; phone: string }) =>
    call<ClaimResult>("/api/claim", { method: "POST", body: JSON.stringify(input) }),

  campaign: () => call<Campaign>("/api/campaign"),

  register: (input: { phone: string; name: string; email?: string; ig_handle?: string; dob?: string }) =>
    call<RegisterResult>("/api/register", { method: "POST", body: JSON.stringify(input) }),

  lookup: (phone: string) =>
    call<LookupResult>("/api/lookup", { method: "POST", body: JSON.stringify({ phone }) }),

  redeem: (input: { phone?: string; code?: string; name?: string; amount?: number }) =>
    call<RedeemResult>("/api/redeem", { method: "POST", body: JSON.stringify(input) }),

  listMembers: () =>
    call<{ data: Member[]; total: number }>("/api/admin/members"),

  listCampaigns: () =>
    call<{ data: unknown[] }>("/api/admin/campaigns"),

  createCampaign: (input: {
    code: string;
    name: string;
    limit: number;
    per_customer?: number;
    discount_type?: "free" | "percent" | "fixed";
    discount_value?: number;
  }) =>
    call<{ campaign: unknown }>("/api/admin/campaigns", {
      method: "POST",
      body: JSON.stringify(input),
    }),
};
