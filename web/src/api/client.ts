interface ApiError {
  error: string;
  message: string;
}

class ApiClient {
  private async request<T>(url: string, options?: RequestInit): Promise<T> {
    const res = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

    if (!res.ok) {
      const body: ApiError = await res.json().catch(() => ({
        error: "unknown",
        message: "Неизвестная ошибка",
      }));
      throw body;
    }

    if (res.status === 204) {
      return undefined as T;
    }

    return res.json();
  }

  requestLink(email: string) {
    return this.request<void>("/api/auth/request-link", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  }

  getMe() {
    return this.request<{ user_id: number; email: string; name: string }>(
      "/api/auth/me"
    );
  }

  logout() {
    return this.request<void>("/api/auth/logout", { method: "POST" });
  }

  createGroup(title: string) {
    return this.request<{ id: number; invite_code: string }>("/api/groups", {
      method: "POST",
      body: JSON.stringify({ title }),
    });
  }

  getGroup(inviteCode: string) {
    return this.request<{
      id: number;
      title: string;
      status: string;
      member_count?: number;
      members?: { name: string; is_me: boolean }[];
      is_organizer?: boolean;
      my_membership_id?: number;
    }>(`/api/groups/${inviteCode}`);
  }

  joinGroup(inviteCode: string, name: string, wishlist: string) {
    return this.request<void>(`/api/groups/${inviteCode}/join`, {
      method: "POST",
      body: JSON.stringify({ name, wishlist }),
    });
  }

  updateWishlist(membershipId: number, wishlist: string) {
    return this.request<void>(`/api/memberships/${membershipId}`, {
      method: "PATCH",
      body: JSON.stringify({ wishlist }),
    });
  }

  draw(groupId: number) {
    return this.request<void>(`/api/groups/${groupId}/draw`, {
      method: "POST",
    });
  }

  getMyRecipient(groupId: number) {
    return this.request<{
      recipient: { name: string; wishlist: string };
    }>(`/api/groups/${groupId}/my-recipient`);
  }

  getChatHistory(groupId: number, role: "santa" | "recipient") {
    return this.request<
      { id: number; from_me: boolean; body: string; created_at: string }[]
    >(`/api/groups/${groupId}/chats/${role}`);
  }
}

export const api = new ApiClient();
